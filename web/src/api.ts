import type { Category, Expense, ExpenseInput, Filter, Member, MemberStatus, Total } from "./types";

const TOKEN_KEY = "expensebot.token";

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
  }
}

export function getToken(): string | null {
  try {
    return localStorage.getItem(TOKEN_KEY);
  } catch {
    return null;
  }
}

function setToken(token: string | null) {
  try {
    if (token) localStorage.setItem(TOKEN_KEY, token);
    else localStorage.removeItem(TOKEN_KEY);
  } catch {
    // приватный режим: вход проживёт до перезагрузки страницы
  }
}

let memoryToken = getToken();
let onUnauthorized: () => void = () => {};

export function setUnauthorizedHandler(handler: () => void) {
  onUnauthorized = handler;
}

// 503 — Cloudflare оборвал запрос по лимиту CPU бесплатного тарифа. GET безопасно повторить;
// запись — нет: она могла успеть сохраниться до обрыва.
const RETRY_DELAYS_MS = [500, 1500];

async function send(method: string, path: string, body?: unknown): Promise<Response> {
  const headers: Record<string, string> = {};
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (memoryToken) headers["Authorization"] = `Bearer ${memoryToken}`;
  const init = { method, headers, body: body === undefined ? undefined : JSON.stringify(body) };
  let resp = await fetch(path, init);
  for (const delay of method === "GET" ? RETRY_DELAYS_MS : []) {
    if (resp.status !== 503) break;
    await new Promise((r) => setTimeout(r, delay));
    resp = await fetch(path, init);
  }
  return resp;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const resp = await send(method, path, body);
  if (resp.status === 204) return undefined as T;
  const data = await resp.json().catch(() => ({}));
  if (!resp.ok) {
    if (resp.status === 401 && path !== "/api/login") {
      logout();
      onUnauthorized();
    }
    const fallback = resp.status === 503 ? "Сервер перегружен, попробуйте ещё раз" : `Ошибка ${resp.status}`;
    throw new ApiError(resp.status, data.error ?? fallback);
  }
  return data as T;
}

export async function login(password: string) {
  const { token } = await request<{ token: string }>("POST", "/api/login", { password });
  memoryToken = token;
  setToken(token);
}

export function logout() {
  memoryToken = null;
  setToken(null);
}

export const PAGE_SIZE = 100;

export function listExpenses(filter: Filter, offset: number, limit = PAGE_SIZE) {
  const q = new URLSearchParams({ limit: String(limit), offset: String(offset) });
  for (const [key, value] of Object.entries(filter)) if (value) q.set(key, value);
  return request<{ items: Expense[]; total: Total }>("GET", `/api/expenses?${q}`);
}

export const listMembers = () => request<Member[]>("GET", "/api/members");
export const updateMember = (id: number, patch: { status?: MemberStatus; name?: string }) =>
  request<Member>("PATCH", `/api/members/${id}`, patch);
export const listCategories = () => request<Category[]>("GET", "/api/categories");
export const createExpense = (input: ExpenseInput) => request<Expense>("POST", "/api/expenses", input);
export const updateExpense = (id: number, input: Partial<ExpenseInput>) =>
  request<Expense>("PATCH", `/api/expenses/${id}`, input);
export const deleteExpense = (id: number) => request<void>("DELETE", `/api/expenses/${id}`);
