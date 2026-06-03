/**
 * AgentCard — static demo widget showing a yaagents agent call trail.
 * v0.3 static; full motion deferred to v0.3.1 (no framer-motion dependency).
 * Uses CSS classes from aimpathy-tokens.css.
 */
interface TrailEntry {
  label: string;
  time: string;
}

interface AgentCardProps {
  name: string;
  role: string;
  status: 'running' | 'done' | 'error' | 'waiting';
  scope: string;
  trail: TrailEntry[];
}

const STATUS_LABELS: Record<AgentCardProps['status'], string> = {
  running: '● running',
  done: '✓ done',
  error: '✗ error',
  waiting: '○ waiting',
};

export default function AgentCard({ name, role, status, scope, trail }: AgentCardProps) {
  return (
    <div className="agent-card-demo" role="region" aria-label={`Agent demo: ${name}`}>
      <div className="acd-header">
        <span>{name}</span>
        <span style={{ fontWeight: 400, fontSize: '0.8rem', color: '#666' }}>({role})</span>
        <span className="acd-status">{STATUS_LABELS[status]}</span>
        <span style={{ marginLeft: 'auto', fontSize: '0.75rem', color: '#999' }}>
          scope: {scope}
        </span>
      </div>
      <ul className="acd-trail" aria-label="Execution trail">
        {trail.map((entry, i) => (
          <li key={i}>
            <span>{entry.label}</span>
            <span className="acd-time">{entry.time}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
