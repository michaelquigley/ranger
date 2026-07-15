import { useCallback, useEffect, useState } from "react";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import {
  fetchItem,
  renameToSlug,
  retitle,
  saveContent,
  type Conflict,
  type ItemDetail,
  type Outcome,
} from "./api";
import { CloseIcon, EditIcon } from "./icons";
import { labelColor, sortedTags } from "./labels";

// the item modal renders the body as markdown by default; the raw-edit
// gesture lives behind an explicit edit mode, where the operator's own
// bytes — frontmatter included — land verbatim through the hash guard. a
// save's conflicts bubble to the board, but a rename's slug_collision
// carries recovery paths the operator needs to see right here.
export function ItemModal({
  filename,
  orderVersion,
  onOutcome,
  onRename,
  onClose,
}: {
  filename: string;
  orderVersion: string;
  onOutcome: (o: Outcome) => boolean;
  onRename: (filename: string) => void;
  onClose: () => void;
}) {
  const [item, setItem] = useState<ItemDetail | null>(null);
  const [editing, setEditing] = useState(false);
  const [editingTitle, setEditingTitle] = useState(false);
  const [content, setContent] = useState("");
  const [title, setTitle] = useState("");
  const [local, setLocal] = useState<string | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const detail = await fetchItem(filename);
      setItem(detail);
      setContent(detail.content);
      setTitle(detail.card.title);
      setLocal(null);
    } catch (err) {
      setLoadError(err instanceof Error ? err.message : String(err));
    }
  }, [filename]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  if (loadError) {
    return (
      <Backdrop onClose={onClose}>
        <div className="panel-head">
          <h2>{filename}</h2>
          <CloseIcon onClick={onClose} />
        </div>
        <p className="local-notice">{loadError}</p>
      </Backdrop>
    );
  }
  if (!item) return null;

  const card = item.card;
  const mismatch = card.flags.some((f) => f.kind === "filename-mismatch");

  const handle = (outcome: Outcome): boolean => {
    if (outcome.kind === "conflict" && outcome.conflict.reason === "slug_collision") {
      setLocal(collisionMessage(outcome.conflict));
      return false;
    }
    if (outcome.kind === "invalid") {
      setLocal(outcome.message);
      return false;
    }
    return onOutcome(outcome);
  };

  const save = async () => {
    if (handle(await saveContent(filename, content, item.hash, orderVersion))) {
      setEditing(false);
      void load();
    }
  };

  const doRetitle = async () => {
    setEditingTitle(false);
    if (title === item.card.title || title === "") {
      setTitle(item.card.title);
      return;
    }
    const outcome = await retitle(filename, title, item.hash, orderVersion);
    if (handle(outcome) && outcome.kind === "ok" && outcome.filename) {
      if (outcome.filename === filename) {
        // an empty-slug retitle keeps the filename; refresh in place
        void load();
      } else {
        onRename(outcome.filename);
      }
    }
  };

  const doRenameToSlug = async () => {
    const outcome = await renameToSlug(filename, item.hash, orderVersion);
    if (handle(outcome) && outcome.kind === "ok" && outcome.filename) {
      onRename(outcome.filename);
    }
  };

  return (
    <Backdrop onClose={onClose}>
      <div className="panel-head">
        {editingTitle ? (
          <input
            className="title-edit"
            value={title}
            autoFocus
            onChange={(e) => setTitle(e.target.value)}
            onBlur={() => {
              setEditingTitle(false);
              setTitle(card.title);
            }}
            onKeyDown={(e) => {
              if (e.key === "Enter") void doRetitle();
              if (e.key === "Escape") {
                e.stopPropagation();
                setEditingTitle(false);
                setTitle(card.title);
              }
            }}
          />
        ) : (
          <h2 className="title-click" title="click to retitle" onClick={() => setEditingTitle(true)}>
            {card.title || filename}
          </h2>
        )}
        <div className="head-actions">
          {!editing && <EditIcon onClick={() => setEditing(true)} />}
          <CloseIcon onClick={onClose} />
        </div>
      </div>
      {local && (
        <p className="local-notice" onClick={() => setLocal(null)}>
          {local}
        </p>
      )}

      <div className="item-meta">
        <span className="meta-pill">{card.state ?? "state unreadable"}</span>
        {card.created && <span className="meta-pill">{card.created}</span>}
        {sortedTags(card.tags).map((tag) => (
          <span key={tag} className="tag-pill" style={labelColor(tag)}>
            {tag}
          </span>
        ))}
        {card.source && <span className="meta-pill meta-source">{card.source}</span>}
        <span className="meta-file">{filename}</span>
      </div>
      {card.flags.length > 0 && (
        <div className="card-flags">
          {card.flags.map((f) => (
            <span key={f.kind} className={`flag flag-${f.kind}`} title={f.diagnostic}>
              {f.kind}: {f.diagnostic}
            </span>
          ))}
        </div>
      )}
      {(card.log ?? []).length > 0 && (
        <div className="item-log">
          {(card.log ?? []).map((entry, i) => (
            <div key={i} className="log-line">
              {entry.stamp} — {entry.note}
            </div>
          ))}
        </div>
      )}

      {editing ? (
        <>
          <textarea
            className="raw-editor"
            value={content}
            onChange={(e) => setContent(e.target.value)}
            spellCheck={false}
          />
          <div className="panel-row">
            <button onClick={() => void save()}>save</button>
            <button
              onClick={() => {
                setContent(item.content);
                setEditing(false);
              }}
            >
              cancel
            </button>
          </div>
        </>
      ) : (
        <>
          <div className="item-body">
            {bodyOf(item.content).trim() ? (
              <Markdown remarkPlugins={[remarkGfm]}>{bodyOf(item.content)}</Markdown>
            ) : (
              <p className="dim">no body.</p>
            )}
          </div>
          {mismatch && (
            <div className="panel-row">
              <button onClick={() => void doRenameToSlug()}>rename to slug</button>
            </div>
          )}
        </>
      )}
    </Backdrop>
  );
}

function Backdrop({ children, onClose }: { children: React.ReactNode; onClose: () => void }) {
  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal modal-item" onClick={(e) => e.stopPropagation()}>
        {children}
      </div>
    </div>
  );
}

// bodyOf splits the frontmatter fence the same way the document layer does:
// opening --- line, closing --- or ... line, body is everything after.
export function bodyOf(content: string): string {
  const lines = content.split("\n");
  if (lines[0]?.trimEnd() !== "---") return content;
  for (let i = 1; i < lines.length; i++) {
    const t = lines[i].trimEnd();
    if (t === "---" || t === "...") {
      return lines.slice(i + 1).join("\n");
    }
  }
  return content;
}

function collisionMessage(conflict: Conflict): string {
  const parts = [conflict.message];
  if (conflict.sourcePath) parts.push(`source: ${conflict.sourcePath}`);
  if (conflict.destPath) parts.push(`collides with: ${conflict.destPath}`);
  return parts.join(" — ");
}
