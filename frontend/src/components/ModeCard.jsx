import React from 'react';

export default function ModeCard({
  title,
  description,
  badges = [],
  children,
  actions = [],
  footer,
}) {
  return (
    <div className="card mode-card">
      <header className="mode-card__header">
        <div>
          <h2>{title}</h2>
          {description && <p>{description}</p>}
        </div>
      </header>

      {badges.length > 0 && (
        <div className="mode-metrics">
          {badges.map((badge, idx) => (
            <span className="badge" key={badge.key ?? idx}>
              {badge.content ?? `${badge.label}: ${badge.value}`}
            </span>
          ))}
        </div>
      )}

      {children}

      {actions.length > 0 && (
        <div className="mode-actions">
          {actions.map((action, idx) => (
            <button
              key={action.key ?? idx}
              className={`primary ${action.variant === 'secondary' ? 'secondary' : ''}`}
              onClick={action.onClick}
              disabled={action.disabled}
            >
              {action.label}
            </button>
          ))}
        </div>
      )}

      {footer && <div className="card-note">{footer}</div>}
    </div>
  );
}
