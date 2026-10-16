export type Expense = {
  id: number;
  user_id: number;
  amount: number; // рубли
  category: string;
  spent_at: string; // YYYY-MM-DD
  comment: string;
  created_at: string;
  updated_at: string;
};

export type MemberStatus = "pending" | "approved" | "rejected";

export type Member = {
  user_id: number;
  username: string;
  first_name: string;
  name: string; // задаёт глава семьи
  display_name: string;
  status: MemberStatus;
  created_at: string;
};

export type Category = { code: string; label: string };

export type Total = {
  count: number;
  amount: number;
  by_category: Record<string, number>;
};

export type Filter = {
  user_id: string;
  from: string;
  to: string;
  category: string;
};

export type ExpenseInput = {
  user_id: number;
  amount: number;
  category: string;
  spent_at: string;
  comment: string;
};
