import { useDroppable } from "@dnd-kit/core";
import { SortableContext, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import type { Card, Lane } from "./api";
import { labelColor, sortedTags } from "./labels";

export function LaneColumn({
  lane,
  onOpen,
  onToggleTag,
  onToggleMilestone,
}: {
  lane: Lane;
  onOpen: (filename: string) => void;
  onToggleTag?: (tag: string) => void;
  onToggleMilestone?: (milestone: string) => void;
}) {
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
              <CardView
                card={card}
                onOpen={onOpen}
                onToggleTag={onToggleTag}
                onToggleMilestone={onToggleMilestone}
              />
            </div>
          ))}
        </div>
      </SortableContext>
    </div>
  );
}

function CardView({
  card,
  onOpen,
  onToggleTag,
  onToggleMilestone,
}: {
  card: Card;
  onOpen: (filename: string) => void;
  onToggleTag?: (tag: string) => void;
  onToggleMilestone?: (milestone: string) => void;
}) {
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
      <CardBody card={card} onToggleTag={onToggleTag} onToggleMilestone={onToggleMilestone} />
    </div>
  );
}

// CardBody is the card's presentation alone — shared by the sortable card
// in its lane and the drag overlay that follows the pointer. tag chips and
// the milestone badge toggle the board filter when handlers are given.
export function CardBody({
  card,
  onToggleTag,
  onToggleMilestone,
}: {
  card: Card;
  onToggleTag?: (tag: string) => void;
  onToggleMilestone?: (milestone: string) => void;
}) {
  return (
    <>
      <div className="card-title">{card.title || card.filename}</div>
      {(card.tags ?? []).length > 0 && (
        <div className="card-tags">
          {sortedTags(card.tags).map((tag) => (
            <span
              key={tag}
              className={onToggleTag ? "tag-pill tag-click" : "tag-pill"}
              style={labelColor(tag)}
              title={onToggleTag ? `filter by ${tag}` : undefined}
              onClick={
                onToggleTag
                  ? (e) => {
                      e.stopPropagation();
                      onToggleTag(tag);
                    }
                  : undefined
              }
            >
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
      {card.milestone && (
        <div
          className={onToggleMilestone ? "card-milestone tag-click" : "card-milestone"}
          title={onToggleMilestone ? `filter by ${card.milestone}` : undefined}
          onClick={
            onToggleMilestone
              ? (e) => {
                  e.stopPropagation();
                  onToggleMilestone(card.milestone!);
                }
              : undefined
          }
        >
          {card.milestone}
        </div>
      )}
    </>
  );
}
