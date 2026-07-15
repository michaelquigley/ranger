import { useCallback, useEffect, useState } from "react";
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  closestCorners,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragOverEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import { fetchBoard, transition, reorder, type Board, type Card, type Conflict, type Outcome, type State } from "./api";
import { rankedAfterMove } from "./reorder";
import { CardBody, LaneColumn } from "./LaneColumn";
import { ItemModal } from "./ItemModal";
import { CaptureModal } from "./CaptureModal";
import { CaptureIcon, VaneMark } from "./icons";

export default function App() {
  const [board, setBoard] = useState<Board | null>(null);
  const [fatal, setFatal] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [openItem, setOpenItem] = useState<string | null>(null);
  const [captureOpen, setCaptureOpen] = useState(false);
  const [dragging, setDragging] = useState<Card | null>(null);
  // server truth as of drag start: the optimistic board mutates freely
  // during the drag, and this is what gestures compute against and what a
  // canceled or failed drop restores.
  const [preDrag, setPreDrag] = useState<Board | null>(null);

  const reload = useCallback(async () => {
    try {
      setBoard(await fetchBoard());
      setFatal(null);
    } catch (err) {
      setFatal(err instanceof Error ? err.message : String(err));
    }
  }, []);

  useEffect(() => {
    void reload();
  }, [reload]);

  // the shared outcome path for gestures resolved at board level:
  // item_conflict and order_conflict mean the view went stale — reload
  // genuinely is the answer. slug_collision and validation belong to the
  // modal that owns the gesture; a fault's message shows verbatim.
  const applyOutcome = useCallback(
    (outcome: Outcome): boolean => {
      switch (outcome.kind) {
        case "ok":
          setBoard(outcome.board);
          setNotice(null);
          return true;
        case "conflict":
          setNotice(conflictNotice(outcome.conflict));
          void reload();
          return false;
        case "invalid":
        case "fault":
          setNotice(outcome.message);
          return false;
      }
    },
    [reload],
  );

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 4 } }));

  const handleDragStart = useCallback(
    (event: DragStartEvent) => {
      if (!board) return;
      const source = locate(board, String(event.active.id));
      setDragging(source ? source.lane.cards[source.index] : null);
      setPreDrag(board);
    },
    [board],
  );

  // live re-parenting: crossing into another lane moves the card there in
  // the optimistic board, so the destination's sortable context opens a
  // slot under the pointer.
  const handleDragOver = useCallback(
    (event: DragOverEvent) => {
      if (!board || !event.over) return;
      const src = locate(board, String(event.active.id));
      const tgt = locateTarget(board, String(event.over.id));
      if (!src || !tgt || src.lane.state === tgt.lane.state) return;
      setBoard(moveAcross(board, src, tgt));
    },
    [board],
  );

  // a failed drop must not leave the optimistic arrangement lying: restore
  // the pre-drag truth and reload (a partial two-file failure may have
  // landed some of it).
  const finishDrag = useCallback(
    (outcome: Outcome, snapshot: Board) => {
      if (!applyOutcome(outcome) && outcome.kind !== "conflict") {
        setBoard(snapshot);
        void reload();
      }
    },
    [applyOutcome, reload],
  );

  const handleDragEnd = useCallback(
    async (event: DragEndEvent) => {
      setDragging(null);
      const snapshot = preDrag;
      setPreDrag(null);
      if (!board || !snapshot) return;
      if (!event.over) {
        setBoard(snapshot);
        return;
      }
      const activeId = String(event.active.id);
      const overId = String(event.over.id);

      const origin = locate(snapshot, activeId);
      const cur = locate(board, activeId);
      if (!origin || !cur) {
        setBoard(snapshot);
        return;
      }

      if (cur.lane.state === origin.lane.state) {
        const tgt = locateTarget(board, overId);
        if (!tgt) {
          setBoard(snapshot);
          return;
        }
        if (tgt.lane.state === origin.lane.state) {
          // within-lane sort
          const filenames = rankedAfterMove(origin.lane, origin.index, tgt.index);
          if (!filenames) {
            setBoard(snapshot);
            return;
          }
          setBoard(optimisticRanked(board, origin.lane.state, filenames));
          finishDrag(await reorder(origin.lane.state as State, filenames, snapshot.orderVersion), snapshot);
          return;
        }
        // dropped into another lane without a drag-over pass
        const position = Math.min(tgt.index, tgt.lane.rankedCount);
        setBoard(moveAcross(board, cur, tgt));
        finishDrag(
          await transition(activeId, tgt.lane.state as State, origin.lane.cards[origin.index].hash, snapshot.orderVersion, position),
          snapshot,
        );
        return;
      }

      // cross-lane: the card already lives in its destination lane; refine
      // the final slot, then transition-and-place. position indexes the
      // destination's ranked list as it exists on disk.
      let finalIndex = cur.index;
      const tgt = locateTarget(board, overId);
      if (tgt && tgt.lane.state === cur.lane.state) {
        finalIndex = Math.min(tgt.index, cur.lane.cards.length - 1);
        if (finalIndex !== cur.index) {
          setBoard(withinLaneMove(board, cur.lane.state, cur.index, finalIndex));
        }
      }
      const diskLane = snapshot.lanes.find((l) => l.state === cur.lane.state);
      const position = Math.min(finalIndex, diskLane?.rankedCount ?? 0);
      finishDrag(
        await transition(activeId, cur.lane.state as State, origin.lane.cards[origin.index].hash, snapshot.orderVersion, position),
        snapshot,
      );
    },
    [board, preDrag, finishDrag],
  );

  const handleDragCancel = useCallback(() => {
    setDragging(null);
    if (preDrag) setBoard(preDrag);
    setPreDrag(null);
  }, [preDrag]);

  if (fatal) {
    return (
      <div className="fatal">
        <h1>vane</h1>
        <p>{fatal}</p>
      </div>
    );
  }
  if (!board) {
    return <div className="fatal">loading…</div>;
  }

  return (
    <div className="app">
      <header>
        <span className="mark" title="vane">
          <VaneMark />
        </span>
        <CaptureIcon onClick={() => setCaptureOpen(true)} />
      </header>
      {notice && (
        <div className="notice" onClick={() => setNotice(null)}>
          {notice}
        </div>
      )}
      <DndContext
        sensors={sensors}
        collisionDetection={closestCorners}
        onDragStart={handleDragStart}
        onDragOver={handleDragOver}
        onDragEnd={handleDragEnd}
        onDragCancel={handleDragCancel}
      >
        <div className="board">
          {board.lanes.map((lane) => (
            <LaneColumn key={lane.state} lane={lane} onOpen={setOpenItem} />
          ))}
        </div>
        <DragOverlay>
          {dragging && (
            <div className="card card-overlay">
              <CardBody card={dragging} />
            </div>
          )}
        </DragOverlay>
      </DndContext>
      {openItem && (
        <ItemModal
          filename={openItem}
          orderVersion={board.orderVersion}
          onOutcome={applyOutcome}
          onRename={(filename) => setOpenItem(filename)}
          onClose={() => setOpenItem(null)}
        />
      )}
      {captureOpen && <CaptureModal onOutcome={applyOutcome} onClose={() => setCaptureOpen(false)} />}
    </div>
  );
}

function conflictNotice(conflict: Conflict): string {
  if (conflict.reason === "item_conflict" || conflict.reason === "order_conflict") {
    return `changed on disk — reloaded (${conflict.message})`;
  }
  return conflict.message;
}

type Located = { lane: Board["lanes"][number]; index: number };

function locate(board: Board, filename: string): Located | null {
  for (const lane of board.lanes) {
    const index = lane.cards.findIndex((c) => c.filename === filename);
    if (index >= 0) return { lane, index };
  }
  return null;
}

// a drop target is either a card (its display index) or a lane container
// (the end of the lane).
function locateTarget(board: Board, overId: string): Located | null {
  if (overId.startsWith("lane:")) {
    const state = overId.slice("lane:".length);
    const lane = board.lanes.find((l) => l.state === state);
    return lane ? { lane, index: lane.cards.length } : null;
  }
  return locate(board, overId);
}

function cloneLanes(board: Board): Board {
  return { ...board, lanes: board.lanes.map((l) => ({ ...l, cards: [...l.cards] })) };
}

// moveAcross relocates a card between lanes in the optimistic board,
// keeping each lane's ranked boundary plausible for preview.
function moveAcross(board: Board, src: Located, tgt: Located): Board {
  const next = cloneLanes(board);
  const from = next.lanes.find((l) => l.state === src.lane.state);
  const to = next.lanes.find((l) => l.state === tgt.lane.state);
  if (!from || !to) return board;
  const [card] = from.cards.splice(src.index, 1);
  if (!card) return board;
  if (src.index < from.rankedCount) from.rankedCount--;
  const at = Math.min(tgt.index, to.cards.length);
  to.cards.splice(at, 0, card);
  if (at <= to.rankedCount) to.rankedCount++;
  return next;
}

function withinLaneMove(board: Board, laneState: string, from: number, to: number): Board {
  const next = cloneLanes(board);
  const lane = next.lanes.find((l) => l.state === laneState);
  if (!lane) return board;
  const [card] = lane.cards.splice(from, 1);
  if (!card) return board;
  lane.cards.splice(Math.min(to, lane.cards.length), 0, card);
  return next;
}

// optimisticRanked repaints one lane as the ranked prefix we just asked the
// server for, with the remaining cards trailing in display order.
function optimisticRanked(board: Board, laneState: string, filenames: string[]): Board {
  const next = cloneLanes(board);
  const lane = next.lanes.find((l) => l.state === laneState);
  if (!lane) return board;
  const byName = new Map(lane.cards.map((c) => [c.filename, c]));
  const ranked: Card[] = [];
  for (const f of filenames) {
    const card = byName.get(f);
    if (card) ranked.push(card);
  }
  const rest = lane.cards.filter((c) => !filenames.includes(c.filename));
  lane.cards = [...ranked, ...rest];
  lane.rankedCount = ranked.length;
  return next;
}
