export type TransactionType = 'income' | 'expense';
export type BudgetPeriod = 'daily' | 'weekly' | 'monthly';

export interface User {
	name: string;
	email: string;
}

export interface LoginResponse {
	token: string;
	user: User;
}

export interface Category {
	id: number;
	name: string;
}

export interface Transaction {
	id: number;
	type: TransactionType;
	category: string;
	amount: number;
	date: string;
	note: string;
}

export interface BudgetRule {
	id: number;
	category: string;
	period: BudgetPeriod;
	limit: number;
}

export interface BudgetCheck {
	category: string;
	period: BudgetPeriod;
	limit: number;
	used: number;
	remaining: number;
	percentage: number;
	over_budget: boolean;
}

export interface TransactionMutationResponse {
	transaction: Transaction;
	budget_checks: BudgetCheck[];
	warning: boolean;
}

export interface DashboardSummary {
	current_balance: number;
	today_expense: number;
	month_expense: number;
	remaining_budget: number;
	makan_today?: BudgetCheck;
	bensin_month?: BudgetCheck;
	budget_usage: BudgetCheck[];
}

export interface MonthlyCategorySpending {
	category: string;
	total: number;
}

export interface MonthlyReport {
	month: string;
	total_income: number;
	total_expense: number;
	net: number;
	spending_by_category: MonthlyCategorySpending[];
	budget_usage: BudgetCheck[];
}
