// Generate PWA icons (PNG) from static/icon.svg using sharp.
// Run with: node scripts/generate-icons.mjs
import { readFile, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import sharp from 'sharp';

const here = path.dirname(fileURLToPath(import.meta.url));
const staticDir = path.resolve(here, '..', 'static');

const sizes = [72, 96, 128, 144, 152, 192, 384, 512];
const extras = [
	{ size: 180, name: 'apple-touch-icon.png' },
	{ size: 64, name: 'favicon.png' },
];

const svg = await readFile(path.join(staticDir, 'icon.svg'));

for (const size of sizes) {
	const out = path.join(staticDir, `icon-${size}x${size}.png`);
	const buf = await sharp(svg, { density: 384 })
		.resize(size, size, { fit: 'contain', background: { r: 0, g: 0, b: 0, alpha: 0 } })
		.png({ compressionLevel: 9 })
		.toBuffer();
	await writeFile(out, buf);
	console.log('wrote', path.basename(out));
}

for (const { size, name } of extras) {
	const out = path.join(staticDir, name);
	const buf = await sharp(svg, { density: 384 })
		.resize(size, size, { fit: 'contain', background: { r: 0, g: 0, b: 0, alpha: 0 } })
		.png({ compressionLevel: 9 })
		.toBuffer();
	await writeFile(out, buf);
	console.log('wrote', name);
}
