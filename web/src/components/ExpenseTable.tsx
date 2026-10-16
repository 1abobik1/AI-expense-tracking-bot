import type { Category, Expense, Member } from "../types";
import { formatDate, formatMoney } from "../format";

type Props = {
  items: Expense[];
  members: Member[];
  categories: Category[];
  onEdit: (e: Expense) => void;
  onDelete: (e: Expense) => void;
};

export default function ExpenseTable({ items, members, categories, onEdit, onDelete }: Props) {
  const memberName = new Map(members.map((m) => [m.user_id, m.display_name]));
  const categoryLabel = new Map(categories.map((c) => [c.code, c.label]));

  if (items.length === 0) return <p className="card empty">Трат по этим условиям нет.</p>;

  return (
    <div className="card table-wrap">
      <table>
        <thead>
          <tr>
            <th>Дата</th>
            <th>Кто</th>
            <th>Категория</th>
            <th>Комментарий</th>
            <th className="num">Сумма</th>
            <th aria-label="Действия" />
          </tr>
        </thead>
        <tbody>
          {items.map((e) => (
            <tr key={e.id}>
              <td className="date">{formatDate(e.spent_at)}</td>
              <td className="who">{memberName.get(e.user_id) ?? e.user_id}</td>
              <td className="category">
                <span className={`tag tag-${e.category}`}>{categoryLabel.get(e.category) ?? e.category}</span>
              </td>
              <td className="comment">{e.comment}</td>
              <td className="num">{formatMoney(e.amount)}</td>
              <td className="actions">
                <button onClick={() => onEdit(e)}>Изменить</button>
                <button className="danger" onClick={() => onDelete(e)}>
                  Удалить
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
