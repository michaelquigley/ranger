import { useState } from "react";
import { type Api, type Outcome } from "./api";

// capture keeps everything typed on screen through every refusal: a
// validation message shows inline (nothing was written), and a
// slug_collision shows the preserved draft's path beside the colliding
// destination, so the operator can retitle and retry knowing exactly where
// the words live.
export function CaptureModal({
  api,
  onOutcome,
  onClose,
}: {
  api: Api;
  onOutcome: (o: Outcome) => boolean;
  onClose: () => void;
}) {
  const [title, setTitle] = useState("");
  const [body, setBody] = useState("");
  const [local, setLocal] = useState<string | null>(null);

  const submit = async () => {
    const outcome = await api.capture(title, body);
    if (outcome.kind === "invalid") {
      setLocal(outcome.message);
      return;
    }
    if (outcome.kind === "conflict") {
      const parts = [outcome.conflict.message];
      if (outcome.conflict.tempPath) parts.push(`draft preserved at: ${outcome.conflict.tempPath}`);
      if (outcome.conflict.destPath) parts.push(`collides with: ${outcome.conflict.destPath}`);
      setLocal(parts.join(" — "));
      return;
    }
    if (onOutcome(outcome)) {
      onClose();
    }
  };

  return (
    <div className="modal-backdrop">
      <div className="modal modal-item">
        <h2>capture</h2>
        {local && <p className="local-notice">{local}</p>}
        <input
          placeholder="title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          autoFocus
        />
        <textarea
          className="capture-body"
          placeholder="body (optional)"
          value={body}
          onChange={(e) => setBody(e.target.value)}
        />
        <div className="panel-row">
          <button onClick={() => void submit()}>capture</button>
          <button onClick={onClose}>cancel</button>
        </div>
      </div>
    </div>
  );
}
