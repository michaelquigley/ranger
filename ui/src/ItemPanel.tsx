import { useCallback, useEffect, useState } from "react";
import {
  fetchItem,
  renameToSlug,
  retitle,
  saveContent,
  type Conflict,
  type ItemDetail,
  type Outcome,
} from "./api";

// the item panel owns the gestures whose refusals are local: a save's
// conflicts bubble to the board, but a rename's slug_collision carries
// recovery paths the operator needs to see right here.
export function ItemPanel({
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

  if (loadError) {
    return (
      <div className="panel">
        <div className="panel-head">
          <h2>{filename}</h2>
          <button onClick={onClose}>close</button>
        </div>
        <p className="local-notice">{loadError}</p>
      </div>
    );
  }
  if (!item) return null;

  const mismatch = item.card.flags.some((f) => f.kind === "filename-mismatch");

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
      void load();
    }
  };

  const doRetitle = async () => {
    const outcome = await retitle(filename, title, item.hash, orderVersion);
    if (handle(outcome) && outcome.kind === "ok" && outcome.filename) {
      onRename(outcome.filename);
    }
  };

  const doRenameToSlug = async () => {
    const outcome = await renameToSlug(filename, item.hash, orderVersion);
    if (handle(outcome) && outcome.kind === "ok" && outcome.filename) {
      onRename(outcome.filename);
    }
  };

  return (
    <div className="panel">
      <div className="panel-head">
        <h2>{filename}</h2>
        <button onClick={onClose}>close</button>
      </div>
      {local && (
        <p className="local-notice" onClick={() => setLocal(null)}>
          {local}
        </p>
      )}
      <div className="panel-row">
        <input value={title} onChange={(e) => setTitle(e.target.value)} />
        <button onClick={() => void doRetitle()}>retitle</button>
        {mismatch && <button onClick={() => void doRenameToSlug()}>rename to slug</button>}
      </div>
      <textarea value={content} onChange={(e) => setContent(e.target.value)} spellCheck={false} />
      <div className="panel-row">
        <button onClick={() => void save()}>save</button>
      </div>
    </div>
  );
}

function collisionMessage(conflict: Conflict): string {
  const parts = [conflict.message];
  if (conflict.sourcePath) parts.push(`source: ${conflict.sourcePath}`);
  if (conflict.destPath) parts.push(`collides with: ${conflict.destPath}`);
  return parts.join(" — ");
}
