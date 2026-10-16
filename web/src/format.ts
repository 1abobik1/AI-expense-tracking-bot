const money = new Intl.NumberFormat("ru-RU", {
  style: "currency",
  currency: "RUB",
  maximumFractionDigits: 2,
});
const date = new Intl.DateTimeFormat("ru-RU", { day: "numeric", month: "short", year: "numeric" });

export function countExpenses(n: number): string {
  const forms = { one: "трата", few: "траты", many: "трат", other: "траты" } as const;
  return `${n} ${forms[new Intl.PluralRules("ru-RU").select(n) as keyof typeof forms]}`;
}

export const formatMoney = (rubles: number) => money.format(rubles);

// Разбираем как местную полночь, иначе дата сдвинется на часовой пояс.
export const formatDate = (iso: string) => {
  const [y, m, d] = iso.split("-").map(Number);
  return date.format(new Date(y, m - 1, d));
};

export const isoDate = (d: Date) =>
  `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;

export const today = () => isoDate(new Date());

export function monthRange(shift: number): { from: string; to: string } {
  const now = new Date();
  const first = new Date(now.getFullYear(), now.getMonth() + shift, 1);
  const last = shift === 0 ? now : new Date(now.getFullYear(), now.getMonth() + shift + 1, 0);
  return { from: isoDate(first), to: isoDate(last) };
}
