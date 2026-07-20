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
// Front three-quarter view: ring objectives, barrels receding up-inward,
// hinge wheel. The masks carve the separation gaps so the drawing stays
// legible on any background.
export function RangerMark() {
  return (
    <svg viewBox="0 0 24 24" width="1.5em" height="1.5em" aria-hidden="true">
      <defs>
        <mask id="rangermark-body">
          <rect width="24" height="24" fill="white" />
          <circle cx="6.6" cy="15.2" r="5.95" fill="black" />
          <circle cx="17.4" cy="15.2" r="5.95" fill="black" />
          <circle cx="12" cy="12.1" r="2.7" fill="black" />
        </mask>
        <mask id="rangermark-rings">
          <rect width="24" height="24" fill="white" />
          <circle cx="12" cy="12.1" r="2.7" fill="black" />
        </mask>
      </defs>
      <g fill="#2563eb" mask="url(#rangermark-body)">
        <path d="M3.83,10.98 L7.09,4.06 A1.7,1.7 0 0 1 10.31,5.14 L8.77,12.62 Z" />
        <path d="M20.17,10.98 L16.91,4.06 A1.7,1.7 0 0 0 13.69,5.14 L15.23,12.62 Z" />
        <path d="M10.75,6.2 A1.25,1.25 0 0 1 13.25,6.2 L13.05,12.2 L10.95,12.2 Z" />
      </g>
      <g fill="none" stroke="#2563eb" mask="url(#rangermark-rings)">
        <circle cx="6.6" cy="15.2" r="4.3" strokeWidth="2.2" />
        <circle cx="17.4" cy="15.2" r="4.3" strokeWidth="2.2" />
      </g>
      <circle cx="12" cy="12.1" r="1.5" fill="none" stroke="#2563eb" strokeWidth="1.2" />
      <g fill="none" stroke="#2563eb" strokeWidth="0.8" strokeLinecap="round">
        <path d="M4.11,14.98 A2.5,2.5 0 0 1 6.38,12.71" />
        <path d="M14.91,14.98 A2.5,2.5 0 0 1 17.18,12.71" />
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
