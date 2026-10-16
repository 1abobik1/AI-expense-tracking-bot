import type { Category, Total } from "../types";
import { countExpenses, formatMoney } from "../format";

export default function Summary({ total, categories }: { total: Total; categories: Category[] }) {
  const rows = categories
    .map((c) => ({ ...c, amount: total.by_category[c.code] ?? 0 }))
    .filter((c) => c.amount > 0)
    .sort((a, b) => b.amount - a.amount);
  const maxAmount = rows[0]?.amount ?? 0;

  return (
    <section className="card summary">
      <div className="summary-total">
        <span className="muted">Итого</span>
        <strong>{formatMoney(total.amount)}</strong>
        <span className="muted">{countExpenses(total.count)}</span>
      </div>
      <ul className="bars">
        {rows.map((r) => (
          <li key={r.code}>
            <span className="bar-label">{r.label}</span>
            <span className="bar-track">
              <span className="bar-fill" style={{ width: `${(r.amount / maxAmount) * 100}%` }} />
            </span>
            <span className="bar-value">{formatMoney(r.amount)}</span>
          </li>
        ))}
      </ul>
    </section>
  );
}
