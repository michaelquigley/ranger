import { useCallback, useEffect, useState } from "react";
import {
  DndContext,
  PointerSensor,
  closestCorners,
  useSensor,
  useSensors,
  type DragEndEvent,
} from "@dnd-kit/core";
import { fetchBoard, transition, reorder, type Board, type Conflict, type Outcome, type State } from "./api";
import { rankedAfterMove } from "./reorder";
import { LaneColumn } from "./LaneColumn";
import { ItemPanel } from "./ItemPanel";
import { CaptureModal } from "./CaptureModal";

export default function App() {
  const [board, setBoard] = useState<Board | null>(null);
  const [fatal, setFatal] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [openItem, setOpenItem] = useState<string | null>(null);
  const [captureOpen, setCaptureOpen] = useState(false);

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
  // modal or panel that owns the gesture; a fault's message shows verbatim.
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

  const handleDragEnd = useCallback(
    async (event: DragEndEvent) => {
      if (!board || !event.over) return;
      const activeId = String(event.active.id);
      const overId = String(event.over.id);
      if (activeId === overId) return;

      const source = locate(board, activeId);
      if (!source) return;
      const target = locateTarget(board, overId);
      if (!target) return;

      if (source.lane.state === target.lane.state) {
        const filenames = rankedAfterMove(source.lane, source.index, target.index);
        if (!filenames) return;
        applyOutcome(await reorder(source.lane.state, filenames, board.orderVersion));
        return;
      }

      // cross-lane: transition-and-place. the position indexes the
      // destination's ranked list only — a drop in the unranked tail
      // serializes as end-of-ranked-list.
      const position = Math.min(target.index, target.lane.rankedCount);
      const card = source.lane.cards[source.index];
      applyOutcome(
        await transition(card.filename, target.lane.state as State, card.hash, board.orderVersion, position),
      );
    },
    [board, applyOutcome],
  );

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
        <h1>vane</h1>
        <button onClick={() => setCaptureOpen(true)}>capture</button>
      </header>
      {notice && (
        <div className="notice" onClick={() => setNotice(null)}>
          {notice}
        </div>
      )}
      <DndContext sensors={sensors} collisionDetection={closestCorners} onDragEnd={handleDragEnd}>
        <div className="board">
          {board.lanes.map((lane) => (
            <LaneColumn key={lane.state} lane={lane} onOpen={setOpenItem} />
          ))}
        </div>
      </DndContext>
      {openItem && (
        <ItemPanel
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
