const TOKEN_KEY = 'finance_token';
const ROLE_KEY = 'finance_role';
const EMAIL_KEY = 'finance_email';
const NAME_KEY = 'finance_name';

export function getAuthToken(): string | null {
	if (typeof localStorage === 'undefined') {
		return null;
	}
	return localStorage.getItem(TOKEN_KEY);
}

export function setAuthToken(token: string): void {
	if (typeof localStorage === 'undefined') {
		return;
	}
	localStorage.setItem(TOKEN_KEY, token);
}

export function clearAuthToken(): void {
	if (typeof localStorage === 'undefined') {
		return;
	}
	localStorage.removeItem(TOKEN_KEY);
	localStorage.removeItem(ROLE_KEY);
	localStorage.removeItem(EMAIL_KEY);
	localStorage.removeItem(NAME_KEY);
}

export function getAuthRole(): string | null {
	if (typeof localStorage === 'undefined') {
		return null;
	}
	return localStorage.getItem(ROLE_KEY);
}

export function setAuthRole(role: string): void {
	if (typeof localStorage === 'undefined') {
		return;
	}
	localStorage.setItem(ROLE_KEY, role);
}

export function isAdmin(): boolean {
	return getAuthRole() === 'admin';
}

export function getAuthEmail(): string | null {
	if (typeof localStorage === 'undefined') return null;
	return localStorage.getItem(EMAIL_KEY);
}

export function setAuthEmail(email: string): void {
	if (typeof localStorage === 'undefined') return;
	localStorage.setItem(EMAIL_KEY, email);
}

export function getAuthName(): string | null {
	if (typeof localStorage === 'undefined') return null;
	return localStorage.getItem(NAME_KEY);
}

export function setAuthName(name: string): void {
	if (typeof localStorage === 'undefined') return;
	localStorage.setItem(NAME_KEY, name);
}
