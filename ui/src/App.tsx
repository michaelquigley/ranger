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
import { fetchBoard, search, transition, reorder, type Board, type Card, type Conflict, type Outcome, type State } from "./api";
import { anchorFor, positionAfterDrop, rankedAfterDrop } from "./reorder";
import { CardBody, LaneColumn } from "./LaneColumn";
import { ItemModal } from "./ItemModal";
import { CaptureModal } from "./CaptureModal";
import { CaptureIcon, RangerMark } from "./icons";
import { labelColor, subsystemColor } from "./labels";

export default function App() {
  const [board, setBoard] = useState<Board | null>(null);
  const [fatal, setFatal] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);
  const [openItem, setOpenItem] = useState<string | null>(null);
  const [captureOpen, setCaptureOpen] = useState(false);
  const [dragging, setDragging] = useState<Card | null>(null);
  const [tagFilter, setTagFilter] = useState<string[]>([]);
  const [subsystemFilter, setSubsystemFilter] = useState<string[]>([]);
  const [milestoneFilter, setMilestoneFilter] = useState<string | null>(null);
  const [query, setQuery] = useState("");
  const [searchMatches, setSearchMatches] = useState<Set<string> | null>(null);
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

  useEffect(() => {
    if (board?.project) {
      document.title = `${board.project} — ranger`;
    }
  }, [board?.project]);

  // debounced title/body search against a fresh server-side read; the
  // match set is one more composable filter.
  useEffect(() => {
    const q = query.trim();
    if (!q) {
      setSearchMatches(null);
      return;
    }
    const timer = setTimeout(async () => {
      try {
        setSearchMatches(new Set(await search(q)));
      } catch (err) {
        setNotice(err instanceof Error ? err.message : String(err));
      }
    }, 150);
    return () => clearTimeout(timer);
  }, [query]);

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
      let working = board;
      let cur = locate(working, activeId);
      if (!origin || !cur) {
        setBoard(snapshot);
        return;
      }

      // a drop can land in a lane the drag never hovered (no re-parent
      // pass); bring the working board up to date first.
      const tgt = locateTarget(working, overId);
      if (tgt && tgt.lane.state !== cur.lane.state) {
        working = moveAcross(working, cur, tgt);
        setBoard(working);
        cur = locate(working, activeId);
        if (!cur) {
          setBoard(snapshot);
          return;
        }
      }

      // placement is anchor-based against the *displayed* lane — under a
      // filter, the card lands directly beside the visible card it was
      // dropped against, and hidden neighbors stay where they are. the
      // gesture itself still computes against the pre-drag server truth.
      const isFiltering =
        tagFilter.length > 0 || subsystemFilter.length > 0 || milestoneFilter !== null || searchMatches !== null;
      const displayedBoard = isFiltering
        ? filterBoard(working, tagFilter, subsystemFilter, milestoneFilter, searchMatches)
        : working;
      const displayedLane = displayedBoard.lanes.find((l) => l.state === cur.lane.state);
      const anchor = anchorFor(displayedLane?.cards.map((c) => c.filename) ?? [], activeId, overId);

      if (cur.lane.state === origin.lane.state) {
        const filenames = rankedAfterDrop(origin.lane, activeId, anchor);
        if (!filenames) {
          setBoard(snapshot);
          return;
        }
        setBoard(optimisticRanked(working, origin.lane.state, filenames));
        finishDrag(await reorder(origin.lane.state as State, filenames, snapshot.orderVersion), snapshot);
        return;
      }

      const destDisk = snapshot.lanes.find((l) => l.state === cur.lane.state);
      const position = destDisk ? positionAfterDrop(destDisk, anchor) : 0;
      // repaint the placement the server will produce before awaiting it —
      // during same-lane hover only dnd-kit transforms move, so without
      // this the release frame shows the card at its lane-entry slot and
      // the lane reshuffles when the response lands.
      setBoard(optimisticPlaced(snapshot, activeId, cur.lane.state, position));
      finishDrag(
        await transition(activeId, cur.lane.state as State, origin.lane.cards[origin.index].hash, snapshot.orderVersion, position),
        snapshot,
      );
    },
    [board, preDrag, finishDrag, tagFilter, subsystemFilter, milestoneFilter, searchMatches],
  );

  const toggleTag = useCallback((tag: string) => {
    setTagFilter((cur) => (cur.includes(tag) ? cur.filter((t) => t !== tag) : [...cur, tag]));
  }, []);

  const toggleSubsystem = useCallback((subsystem: string) => {
    setSubsystemFilter((cur) => (cur.includes(subsystem) ? cur.filter((s) => s !== subsystem) : [...cur, subsystem]));
  }, []);

  const toggleMilestone = useCallback((milestone: string) => {
    setMilestoneFilter((cur) => (cur === milestone ? null : milestone));
  }, []);

  const handleDragCancel = useCallback(() => {
    setDragging(null);
    if (preDrag) setBoard(preDrag);
    setPreDrag(null);
  }, [preDrag]);

  if (fatal) {
    return (
      <div className="fatal">
        <h1>ranger</h1>
        <p>{fatal}</p>
      </div>
    );
  }
  if (!board) {
    return <div className="fatal">loading…</div>;
  }

  // an active filter narrows the board to cards carrying every selected
  // tag and the selected milestone. drags stay live: placement is
  // anchor-based, so a filtered drop lands beside the visible card it was
  // dropped against.
  const filtering =
    tagFilter.length > 0 || subsystemFilter.length > 0 || milestoneFilter !== null || searchMatches !== null;
  const shown = filtering ? filterBoard(board, tagFilter, subsystemFilter, milestoneFilter, searchMatches) : board;

  return (
    <div className="app">
      <header>
        <span className="mark" title="ranger">
          <RangerMark />
        </span>
        <h1 className="project-name">{board.project}</h1>
        <input
          className="search-box"
          placeholder="search"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Escape") setQuery("");
          }}
        />
        <CaptureIcon onClick={() => setCaptureOpen(true)} />
        {filtering && (
          <div className="filter-bar">
            <span className="dim">filtering:</span>
            {milestoneFilter && (
              <span
                className="milestone-pill tag-click"
                title={`stop filtering by ${milestoneFilter}`}
                onClick={() => setMilestoneFilter(null)}
              >
                {milestoneFilter} ×
              </span>
            )}
            {subsystemFilter.map((subsystem) => (
              <span
                key={subsystem}
                className="subsystem-pill tag-click"
                style={subsystemColor(subsystem)}
                title={`stop filtering by ${subsystem}`}
                onClick={() => toggleSubsystem(subsystem)}
              >
                {subsystem} ×
              </span>
            ))}
            {tagFilter.map((tag) => (
              <span
                key={tag}
                className="tag-pill tag-click"
                style={labelColor(tag)}
                title={`stop filtering by ${tag}`}
                onClick={() => toggleTag(tag)}
              >
                {tag} ×
              </span>
            ))}
            <button
              onClick={() => {
                setTagFilter([]);
                setSubsystemFilter([]);
                setMilestoneFilter(null);
              }}
            >
              clear
            </button>
          </div>
        )}
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
          {shown.lanes.map((lane) => (
            <LaneColumn
              key={lane.state}
              lane={lane}
              onOpen={setOpenItem}
              onToggleTag={toggleTag}
              onToggleSubsystem={toggleSubsystem}
              onToggleMilestone={toggleMilestone}
            />
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

// filterBoard narrows every lane to cards carrying all of the given tags
// and the selected milestone. rankedCount shrinks to the surviving members
// of the ranked prefix so the boundary rule still lands between ranked and
// unranked cards.
function filterBoard(
  board: Board,
  tags: string[],
  subsystems: string[],
  milestone: string | null,
  searchMatches: Set<string> | null,
): Board {
  const matches = (c: Card) =>
    tags.every((t) => (c.tags ?? []).includes(t)) &&
    subsystems.every((s) => (c.subsystems ?? []).includes(s)) &&
    (milestone === null || c.milestone === milestone) &&
    (searchMatches === null || searchMatches.has(c.filename));
  return {
    ...board,
    lanes: board.lanes.map((lane) => ({
      ...lane,
      cards: lane.cards.filter(matches),
      rankedCount: lane.cards.slice(0, lane.rankedCount).filter(matches).length,
    })),
  };
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

// optimisticPlaced is the cross-lane sibling of optimisticRanked: the moved
// card leaves its origin lane and enters the destination's ranked prefix at
// position, computed from the pre-drag server truth the gesture was sent
// against.
function optimisticPlaced(snapshot: Board, activeId: string, destState: string, position: number): Board {
  const next = cloneLanes(snapshot);
  const src = locate(next, activeId);
  const to = next.lanes.find((l) => l.state === destState);
  if (!src || !to || src.lane.state === destState) return snapshot;
  const [card] = src.lane.cards.splice(src.index, 1);
  if (!card) return snapshot;
  if (src.index < src.lane.rankedCount) src.lane.rankedCount--;
  to.cards.splice(Math.min(position, to.rankedCount), 0, card);
  to.rankedCount++;
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
