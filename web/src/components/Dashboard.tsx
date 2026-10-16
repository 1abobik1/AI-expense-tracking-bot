import { useCallback, useEffect, useState } from "react";
import * as api from "../api";
import type { Category, Expense, ExpenseInput, Filter, Member, MemberStatus, Total } from "../types";
import { formatMoney, monthRange } from "../format";
import Filters from "./Filters";
import Summary from "./Summary";
import ExpenseTable from "./ExpenseTable";
import ExpenseForm from "./ExpenseForm";
import Requests from "./Requests";
import Family from "./Family";

const emptyTotal: Total = { count: 0, amount: 0, by_category: {} };

export default function Dashboard({ onLogout }: { onLogout: () => void }) {
  const [members, setMembers] = useState<Member[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [filter, setFilter] = useState<Filter>({ user_id: "", category: "", ...monthRange(0) });
  const [items, setItems] = useState<Expense[]>([]);
  const [total, setTotal] = useState<Total>(emptyTotal);
  const [editing, setEditing] = useState<Expense | null | undefined>(undefined); // undefined — форма закрыта
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [familyOpen, setFamilyOpen] = useState(false);
  const [updatedAt, setUpdatedAt] = useState(() => new Date());
  const approved = members.filter((m) => m.status === "approved");
  const pending = members.filter((m) => m.status === "pending");

  useEffect(() => {
    Promise.all([api.listMembers(), api.listCategories()])
      .then(([m, c]) => {
        setMembers(m);
        setCategories(c);
      })
      .catch((e) => setError(e.message));
  }, []);

  const load = useCallback(
    async (offset = 0) => {
      setLoading(true);
      try {
        const page = await api.listExpenses(filter, offset);
        setItems((prev) => (offset === 0 ? page.items : [...prev, ...page.items]));
        setTotal(page.total);
        setUpdatedAt(new Date());
        setError("");
      } catch (e) {
        setError(e instanceof Error ? e.message : "Не удалось загрузить траты");
      } finally {
        setLoading(false);
      }
    },
    [filter],
  );

  useEffect(() => {
    load(0);
  }, [load]);

  // По кнопке, а не по таймеру: каждый запрос тратит CPU бесплатного тарифа Cloudflare.
  async function refresh() {
    setLoading(true);
    try {
      const [page, m] = await Promise.all([
        api.listExpenses(filter, 0, Math.max(items.length, api.PAGE_SIZE)),
        api.listMembers(),
      ]);
      setItems(page.items);
      setTotal(page.total);
      setMembers(m);
      setUpdatedAt(new Date());
      setError("");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Не удалось обновить");
    } finally {
      setLoading(false);
    }
  }

  async function updateMember(m: Member, patch: { status?: MemberStatus; name?: string }) {
    try {
      const updated = await api.updateMember(m.user_id, patch);
      setMembers((prev) => prev.map((x) => (x.user_id === updated.user_id ? updated : x)));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Не удалось сохранить");
    }
  }

  async function save(input: ExpenseInput) {
    if (editing) {
      const { user_id: _, ...patch } = input; // владельца траты не меняем
      await api.updateExpense(editing.id, patch);
    } else {
      await api.createExpense(input);
    }
    setEditing(undefined);
    await load(0);
  }

  async function remove(e: Expense) {
    if (!confirm(`Удалить трату «${e.comment || e.category}» на ${formatMoney(e.amount)}?`)) return;
    try {
      await api.deleteExpense(e.id);
      await load(0);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Не удалось удалить");
    }
  }

  return (
    <div className="layout">
      <header className="topbar">
        <div>
          <h1>Семейные расходы</h1>
          <span className="muted small">обновлено в {updatedAt.toLocaleTimeString("ru-RU")}</span>
        </div>
        <div className="topbar-actions">
          <button className="primary" disabled={approved.length === 0} onClick={() => setEditing(null)}>
            + Добавить трату
          </button>
          <button disabled={loading} onClick={refresh}>
            Обновить
          </button>
          <button onClick={() => setFamilyOpen(true)}>Семья ({approved.length})</button>
          <button onClick={onLogout}>Выйти</button>
        </div>
      </header>

      {members.length === 0 && !error && (
        <p className="card hint">
          Пока никто не писал боту. Заявки появятся здесь после первого сообщения боту.
        </p>
      )}
      {error && <p className="card error">{error}</p>}

      <Requests
        pending={pending}
        onDecide={(m, approve, name) =>
          updateMember(m, approve ? { status: "approved", name } : { status: "rejected" })
        }
      />
      <Filters filter={filter} members={approved} categories={categories} onChange={setFilter} />
      <Summary total={total} categories={categories} />
      <ExpenseTable
        items={items}
        members={members}
        categories={categories}
        onEdit={setEditing}
        onDelete={remove}
      />
      {items.length < total.count && (
        <button className="more" disabled={loading} onClick={() => load(items.length)}>
          {loading ? "Загрузка…" : `Показать ещё (${total.count - items.length})`}
        </button>
      )}

      {editing !== undefined && (
        <ExpenseForm
          expense={editing}
          members={editing ? members : approved}
          categories={categories}
          onSave={save}
          onClose={() => setEditing(undefined)}
        />
      )}
      {familyOpen && (
        <Family members={members} onUpdate={updateMember} onClose={() => setFamilyOpen(false)} />
      )}
    </div>
  );
}
