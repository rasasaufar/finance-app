import { browser } from '$app/environment';
import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

export const ssr = false;
export const prerender = false;

export const load: LayoutLoad = ({ url }) => {
	const pathname = url.pathname;
	const isLoginPage = pathname === '/login';
	const token = browser ? localStorage.getItem('finance_token') : null;
	const isLoggedIn = Boolean(token);

	if (!isLoggedIn && !isLoginPage) {
		throw redirect(302, '/login');
	}

	if (isLoggedIn && isLoginPage) {
		throw redirect(302, '/dashboard');
	}

	return {
		isLoggedIn
	};
};
