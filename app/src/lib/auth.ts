const TOKEN_KEY = 'finance_token';

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
}
