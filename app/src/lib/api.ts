import { PUBLIC_API_BASE_URL } from '$env/static/public';
import { clearAuthToken, getAuthToken } from '$lib/auth';

export class ApiError extends Error {
	status: number;

	constructor(message: string, status: number) {
		super(message);
		this.name = 'ApiError';
		this.status = status;
	}
}

async function request<T>(path: string, options: RequestInit = {}, withAuth = true): Promise<T> {
	const headers = new Headers(options.headers);
	headers.set('Content-Type', 'application/json');

	if (withAuth) {
		const token = getAuthToken();
		if (token) {
			headers.set('Authorization', `Bearer ${token}`);
		}
	}

	const response = await fetch(`${PUBLIC_API_BASE_URL}${path}`, {
		...options,
		headers
	});

	if (response.status === 401 && withAuth) {
		clearAuthToken();
		if (typeof window !== 'undefined') {
			window.location.href = '/login';
		}
	}

	if (response.status === 204) {
		return undefined as T;
	}

	const rawBody = await response.text();
	let payload: unknown = null;

	if (rawBody) {
		try {
			payload = JSON.parse(rawBody);
		} catch {
			payload = null;
		}
	}

	if (!response.ok) {
		const message =
			typeof payload === 'object' && payload !== null && 'error' in payload
				? String((payload as { error: string }).error)
				: 'Terjadi kesalahan saat memproses data.';
		throw new ApiError(message, response.status);
	}

	return payload as T;
}

export const api = {
	get: <T>(path: string, withAuth = true) => request<T>(path, { method: 'GET' }, withAuth),
	post: <T>(path: string, body: unknown, withAuth = true) =>
		request<T>(
			path,
			{
				method: 'POST',
				body: JSON.stringify(body)
			},
			withAuth
		),
	put: <T>(path: string, body: unknown, withAuth = true) =>
		request<T>(
			path,
			{
				method: 'PUT',
				body: JSON.stringify(body)
			},
			withAuth
		),
	delete: <T>(path: string, withAuth = true) => request<T>(path, { method: 'DELETE' }, withAuth)
};
