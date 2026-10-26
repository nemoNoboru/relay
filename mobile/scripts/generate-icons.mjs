// Renders the app icons from assets/icon.svg. Run with `npm run icons`.
import { readFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import sharp from 'sharp'

const assets = join(dirname(fileURLToPath(import.meta.url)), '..', 'assets')
const PAPER = '#f0e6d2'

const source = await readFile(join(assets, 'icon.svg'), 'utf8')
const mark = source.split('<!-- mark:start -->')[1]?.split('<!-- mark:end -->')[0]
if (!mark) throw new Error('icon.svg is missing the mark:start / mark:end markers')

/** The mark, scaled around the canvas centre, on an optional solid background. */
const compose = ({ scale, background, style = '' }) => `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024">
  <style>${style}</style>
  ${background ? `<rect width="1024" height="1024" fill="${background}"/>` : ''}
  <g transform="translate(512 512) scale(${scale}) translate(-512 -512)">${mark}</g>
</svg>`

const render = (svg, size) => sharp(Buffer.from(svg), { density: 384 }).resize(size, size)

const save = (name, pipeline) => pipeline.png().toFile(join(assets, name))

// iOS: opaque, the OS applies the rounded mask.
await save('icon.png', render(compose({ scale: 1.3, background: PAPER }), 1024).flatten({ background: PAPER }))

// Android adaptive: 108dp canvas, only the centre 66% is guaranteed visible.
await save('android-icon-background.png', render(compose({ scale: 0, background: PAPER }), 1024))
await save('android-icon-foreground.png', render(compose({ scale: 0.92 }), 1024))

// Themed icon: alpha-only silhouette, the R knocked out of the square.
const mono = await render(
  compose({
    scale: 0.92,
    style: '.shadow,.accent{display:none}.face{fill:#000}.r{fill:#fff}',
  }),
  1024,
)
  .ensureAlpha()
  .raw()
  .toBuffer({ resolveWithObject: true })
for (let i = 0; i < mono.data.length; i += 4) {
  const knockout = mono.data[i] / 255 // red channel: 0 on the face, 255 on the R
  mono.data[i] = mono.data[i + 1] = mono.data[i + 2] = 0
  mono.data[i + 3] = Math.round(mono.data[i + 3] * (1 - knockout))
}
await save('android-icon-monochrome.png', sharp(mono.data, { raw: mono.info }))

await save('splash-icon.png', render(compose({ scale: 1 }), 1024))
await save('favicon.png', render(compose({ scale: 1.3, background: PAPER }), 256))
