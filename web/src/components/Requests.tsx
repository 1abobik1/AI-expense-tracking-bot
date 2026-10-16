import { useState } from "react";
import type { Member } from "../types";
import { formatDate } from "../format";

type Props = {
  pending: Member[];
  onDecide: (m: Member, approve: boolean, name: string) => Promise<void>;
};

export function telegramLabel(m: Member) {
  const parts = [m.first_name, m.username && `@${m.username}`, `id ${m.user_id}`].filter(Boolean);
  return parts.join(" · ");
}

export default function Requests({ pending, onDecide }: Props) {
  if (pending.length === 0) return null;
  return (
    <section className="card requests">
      <h2>Заявки в семью</h2>
      <p className="muted">Эти люди написали боту. Пока вы их не примете, бот не записывает их траты.</p>
      {pending.map((m) => (
        <Request key={m.user_id} member={m} onDecide={onDecide} />
      ))}
    </section>
  );
}

function Request({ member, onDecide }: { member: Member; onDecide: Props["onDecide"] }) {
  const [name, setName] = useState(member.first_name);
  const [busy, setBusy] = useState(false);
  const decide = async (approve: boolean) => {
    setBusy(true);
    try {
      await onDecide(member, approve, name);
    } finally {
      setBusy(false);
    }
  };
  return (
    <div className="request">
      <div>
        <div>{telegramLabel(member)}</div>
        <div className="muted small">написал {formatDate(member.created_at.slice(0, 10))}</div>
      </div>
      <input
        placeholder="Имя в семье: Мама, Сын…"
        value={name}
        maxLength={50}
        onChange={(e) => setName(e.target.value)}
      />
      <div className="request-actions">
        <button className="primary" disabled={busy} onClick={() => decide(true)}>
          Принять
        </button>
        <button className="danger" disabled={busy} onClick={() => decide(false)}>
          Отклонить
        </button>
      </div>
    </div>
  );
}
