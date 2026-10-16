import { useState } from "react";
import type { Member, MemberStatus } from "../types";
import { telegramLabel } from "./Requests";

type Props = {
  members: Member[];
  onUpdate: (m: Member, patch: { status?: MemberStatus; name?: string }) => Promise<void>;
  onClose: () => void;
};

export default function Family({ members, onUpdate, onClose }: Props) {
  const approved = members.filter((m) => m.status === "approved");
  const rejected = members.filter((m) => m.status === "rejected");
  return (
    <div className="backdrop" onClick={onClose}>
      <div className="card dialog family" onClick={(e) => e.stopPropagation()}>
        <h2>Семья</h2>
        {approved.length === 0 && <p className="muted">Пока никого. Примите заявки тех, кто написал боту.</p>}
        {approved.map((m) => (
          <FamilyMember key={m.user_id} member={m} onUpdate={onUpdate} />
        ))}
        {rejected.length > 0 && (
          <>
            <h3>Без доступа</h3>
            {rejected.map((m) => (
              <div key={m.user_id} className="family-row">
                <span className="family-name">{m.display_name}</span>
                <span className="muted small">{telegramLabel(m)}</span>
                <button onClick={() => onUpdate(m, { status: "approved" })}>Вернуть доступ</button>
              </div>
            ))}
          </>
        )}
        <div className="dialog-actions">
          <button onClick={onClose}>Закрыть</button>
        </div>
      </div>
    </div>
  );
}

function FamilyMember({ member, onUpdate }: { member: Member; onUpdate: Props["onUpdate"] }) {
  const [name, setName] = useState(member.name);
  const changed = name.trim() !== member.name;
  return (
    <div className="family-row">
      <input
        value={name}
        placeholder={member.first_name || "Имя"}
        maxLength={50}
        onChange={(e) => setName(e.target.value)}
      />
      <span className="muted small">{telegramLabel(member)}</span>
      <div className="request-actions">
        {changed && (
          <button className="primary" onClick={() => onUpdate(member, { name })}>
            Сохранить
          </button>
        )}
        <button
          className="danger"
          onClick={() =>
            confirm(`Закрыть доступ для «${member.display_name}»? Траты останутся.`) &&
            onUpdate(member, { status: "rejected" })
          }
        >
          Закрыть доступ
        </button>
      </div>
    </div>
  );
}
