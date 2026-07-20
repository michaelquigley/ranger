// material design icon paths (Apache-2.0), inlined so the binary stays
// self-contained — no icon font, no CDN.

function Icon({ d, label, onClick }: { d: string; label: string; onClick: () => void }) {
  return (
    <button className="icon-button" title={label} aria-label={label} onClick={onClick}>
      <svg viewBox="0 0 24 24" width="1.2em" height="1.2em" fill="currentColor" aria-hidden="true">
        <path d={d} />
      </svg>
    </button>
  );
}

// RangerMark is the binoculars mark — the same drawing as the favicon.
export function RangerMark() {
  return (
    <svg viewBox="0 0 24 24" width="1.5em" height="1.5em" aria-hidden="true">
      <g fill="#2563eb">
        <rect x="4.9" y="3.5" width="4.2" height="7.4" rx="2" />
        <rect x="14.9" y="3.5" width="4.2" height="7.4" rx="2" />
        <rect x="10.6" y="6" width="2.8" height="4" rx="1.2" />
        <circle cx="7" cy="15.3" r="4.6" />
        <circle cx="17" cy="15.3" r="4.6" />
      </g>
    </svg>
  );
}

export function CaptureIcon({ onClick }: { onClick: () => void }) {
  return <Icon label="capture" onClick={onClick} d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" />;
}

export function EditIcon({ onClick }: { onClick: () => void }) {
  return (
    <Icon
      label="edit raw content"
      onClick={onClick}
      d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34a.9959.9959 0 0 0-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"
    />
  );
}

export function DeleteIcon({ onClick }: { onClick: () => void }) {
  return (
    <Icon
      label="delete item"
      onClick={onClick}
      d="M6 19c0 1.1.9 2 2 2h8c1.1 0 2-.9 2-2V7H6v12zM19 4h-3.5l-1-1h-5l-1 1H5v2h14V4z"
    />
  );
}

export function CloseIcon({ onClick }: { onClick: () => void }) {
  return (
    <Icon
      label="close"
      onClick={onClick}
      d="M19 6.41 17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"
    />
  );
}
