import { useState, type FormEvent } from "react";
import type { Category, Expense, ExpenseInput, Member } from "../types";
import { today } from "../format";

type Props = {
  expense: Expense | null; // null — новая трата
  members: Member[];
  categories: Category[];
  onSave: (input: ExpenseInput) => Promise<void>;
  onClose: () => void;
};

export default function ExpenseForm({ expense, members, categories, onSave, onClose }: Props) {
  const [form, setForm] = useState({
    user_id: String(expense?.user_id ?? members[0]?.user_id ?? ""),
    amount: expense ? String(expense.amount) : "",
    category: expense?.category ?? "food",
    spent_at: expense?.spent_at ?? today(),
    comment: expense?.comment ?? "",
  });
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const set = (patch: Partial<typeof form>) => setForm({ ...form, ...patch });

  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await onSave({
        user_id: Number(form.user_id),
        amount: Number(form.amount.replace(",", ".")),
        category: form.category,
        spent_at: form.spent_at,
        comment: form.comment.trim(),
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось сохранить");
      setBusy(false);
    }
  }

  return (
    <div className="backdrop" onClick={onClose}>
      <form className="card dialog" onSubmit={submit} onClick={(e) => e.stopPropagation()}>
        <h2>{expense ? "Изменить трату" : "Новая трата"}</h2>
        <label>
          Член семьи
          <select
            value={form.user_id}
            disabled={expense !== null}
            onChange={(e) => set({ user_id: e.target.value })}
            required
          >
            {members.map((m) => (
              <option key={m.user_id} value={m.user_id}>
                {m.display_name}
              </option>
            ))}
          </select>
        </label>
        <label>
          Сумма, ₽
          <input
            inputMode="decimal"
            value={form.amount}
            onChange={(e) => set({ amount: e.target.value })}
            required
            autoFocus
          />
        </label>
        <label>
          Категория
          <select value={form.category} onChange={(e) => set({ category: e.target.value })}>
            {categories.map((c) => (
              <option key={c.code} value={c.code}>
                {c.label}
              </option>
            ))}
          </select>
        </label>
        <label>
          Дата
          <input
            type="date"
            value={form.spent_at}
            max={today()}
            onChange={(e) => set({ spent_at: e.target.value })}
            required
          />
        </label>
        <label>
          Комментарий
          <input value={form.comment} maxLength={200} onChange={(e) => set({ comment: e.target.value })} />
        </label>
        {error && <p className="error">{error}</p>}
        <div className="dialog-actions">
          <button type="button" onClick={onClose}>
            Отмена
          </button>
          <button className="primary" disabled={busy}>
            Сохранить
          </button>
        </div>
      </form>
    </div>
  );
}
