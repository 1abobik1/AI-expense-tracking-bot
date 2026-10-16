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

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {};
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (memoryToken) headers["Authorization"] = `Bearer ${memoryToken}`;
  const resp = await fetch(path, {
    method,
    headers,
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (resp.status === 204) return undefined as T;
  const data = await resp.json().catch(() => ({}));
  if (!resp.ok) {
    if (resp.status === 401 && path !== "/api/login") {
      logout();
      onUnauthorized();
    }
    throw new ApiError(resp.status, data.error ?? `Ошибка ${resp.status}`);
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
