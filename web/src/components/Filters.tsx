import type { Category, Filter, Member } from "../types";
import { monthRange } from "../format";

type Props = {
  filter: Filter;
  members: Member[];
  categories: Category[];
  onChange: (f: Filter) => void;
};

const presets = [
  { label: "Этот месяц", range: () => monthRange(0) },
  { label: "Прошлый месяц", range: () => monthRange(-1) },
  { label: "Всё время", range: () => ({ from: "", to: "" }) },
];

export default function Filters({ filter, members, categories, onChange }: Props) {
  const set = (patch: Partial<Filter>) => onChange({ ...filter, ...patch });
  return (
    <section className="card filters">
      <label>
        Член семьи
        <select value={filter.user_id} onChange={(e) => set({ user_id: e.target.value })}>
          <option value="">Все</option>
          {members.map((m) => (
            <option key={m.user_id} value={m.user_id}>
              {m.display_name}
            </option>
          ))}
        </select>
      </label>
      <label>
        Категория
        <select value={filter.category} onChange={(e) => set({ category: e.target.value })}>
          <option value="">Все</option>
          {categories.map((c) => (
            <option key={c.code} value={c.code}>
              {c.label}
            </option>
          ))}
        </select>
      </label>
      <label>
        С
        <input type="date" value={filter.from} onChange={(e) => set({ from: e.target.value })} />
      </label>
      <label>
        По
        <input type="date" value={filter.to} onChange={(e) => set({ to: e.target.value })} />
      </label>
      <div className="presets">
        {presets.map((p) => (
          <button key={p.label} type="button" onClick={() => set(p.range())}>
            {p.label}
          </button>
        ))}
      </div>
    </section>
  );
}
