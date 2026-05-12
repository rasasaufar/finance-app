import { browser } from '$app/environment';
import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

export const ssr = false;
export const prerender = false;

export const load: LayoutLoad = ({ url }) => {
	const pathname = url.pathname;
	const isLoginPage = pathname === '/login';
	const token = browser ? localStorage.getItem('finance_token') : null;
	const role = browser ? (localStorage.getItem('finance_role') ?? '') : '';
	const isLoggedIn = Boolean(token);

	if (!isLoggedIn && !isLoginPage) {
		throw redirect(302, '/login');
	}

	if (isLoggedIn && isLoginPage) {
		throw redirect(302, '/dashboard');
	}

	// Guard admin routes
	if (isLoggedIn && pathname.startsWith('/admin') && role !== 'admin') {
		throw redirect(302, '/dashboard');
	}

	return {
		isLoggedIn,
		role
	};
};
