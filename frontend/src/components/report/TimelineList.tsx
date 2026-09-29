// TimelineList — chronological event timeline with raw log evidence.
import type { TimelineEvent } from '../../api/backend';

export interface TimelineListProps {
  events: TimelineEvent[];
}

export function TimelineList({ events }: TimelineListProps) {
  return (
    <div className="timeline">
      {events.map((event, index) => (
        <div className="timeline-event" key={`${event.timestamp}-${index}`}>
          <div className="timeline-time">{event.timestamp}</div>
          <div className="timeline-msg">{event.message}</div>
          {event.raw_log ? <div className="timeline-raw">{event.raw_log}</div> : null}
        </div>
      ))}
    </div>
  );
}

export default TimelineList;
