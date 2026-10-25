// Crop a PNG to its top <height> CSS px (x2 scale). Headless chromium loses
// 87px of viewport to window chrome, so render.sh over-sizes the window and
// trims the blank bottom here. Cropping bottom rows needs no unfiltering.
const fs = require('fs'), zlib = require('zlib');
const [file, cssH] = [process.argv[2], +process.argv[3]];
const buf = fs.readFileSync(file);
let pos = 8, ihdr, idat = [];
while (pos < buf.length) {
  const len = buf.readUInt32BE(pos), type = buf.toString('ascii', pos + 4, pos + 8);
  const data = buf.subarray(pos + 8, pos + 8 + len);
  if (type === 'IHDR') ihdr = Buffer.from(data);
  if (type === 'IDAT') idat.push(data);
  pos += 12 + len;
}
const w = ihdr.readUInt32BE(0), bpp = { 2: 3, 6: 4 }[ihdr[9]];
if (!bpp || ihdr[8] !== 8) throw new Error('unsupported PNG');
const h = Math.round(cssH * (w / +process.argv[4]));
const raw = zlib.inflateSync(Buffer.concat(idat)).subarray(0, h * (1 + w * bpp));
ihdr.writeUInt32BE(h, 4);
const chunk = (t, d) => {
  const b = Buffer.alloc(12 + d.length);
  b.writeUInt32BE(d.length, 0); b.write(t, 4, 'ascii'); d.copy(b, 8);
  b.writeUInt32BE(zlib.crc32(b.subarray(4, 8 + d.length)), 8 + d.length);
  return b;
};
fs.writeFileSync(file, Buffer.concat([buf.subarray(0, 8), chunk('IHDR', ihdr),
  chunk('IDAT', zlib.deflateSync(raw)), chunk('IEND', Buffer.alloc(0))]));
