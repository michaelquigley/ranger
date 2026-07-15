import { useDroppable } from "@dnd-kit/core";
import { SortableContext, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import type { Card, Lane } from "./api";
import { labelColor, sortedTags } from "./labels";

export function LaneColumn({ lane, onOpen }: { lane: Lane; onOpen: (filename: string) => void }) {
  const { setNodeRef } = useDroppable({ id: `lane:${lane.state}` });
  return (
    <div className="lane" ref={setNodeRef}>
      <h2>
        {lane.state} <span className="count">{lane.cards.length}</span>
      </h2>
      <SortableContext items={lane.cards.map((c) => c.filename)} strategy={verticalListSortingStrategy}>
        <div className="cards">
          {lane.cards.map((card, i) => (
            <div key={card.filename}>
              {i === lane.rankedCount && lane.rankedCount > 0 && <div className="rank-boundary" />}
              <CardView card={card} onOpen={onOpen} />
            </div>
          ))}
        </div>
      </SortableContext>
    </div>
  );
}

function CardView({ card, onOpen }: { card: Card; onOpen: (filename: string) => void }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: card.filename,
  });
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.35 : undefined,
  };
  return (
    <div
      className="card"
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      onClick={() => onOpen(card.filename)}
    >
      <CardBody card={card} />
    </div>
  );
}

// CardBody is the card's presentation alone — shared by the sortable card
// in its lane and the drag overlay that follows the pointer.
export function CardBody({ card }: { card: Card }) {
  return (
    <>
      <div className="card-title">{card.title || card.filename}</div>
      {(card.tags ?? []).length > 0 && (
        <div className="card-tags">
          {sortedTags(card.tags).map((tag) => (
            <span key={tag} className="tag-pill" style={labelColor(tag)}>
              {tag}
            </span>
          ))}
        </div>
      )}
      {card.flags.length > 0 && (
        <div className="card-flags">
          {card.flags.map((f) => (
            <span key={f.kind} className={`flag flag-${f.kind}`} title={f.diagnostic}>
              {f.kind}
            </span>
          ))}
        </div>
      )}
      {(card.log ?? []).map((entry, i) => (
        <div key={i} className="log-line">
          {entry.stamp} — {entry.note}
        </div>
      ))}
    </>
  );
}
