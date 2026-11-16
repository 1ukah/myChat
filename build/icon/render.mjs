// Renders every icon file from the SVG masters. Each size is drawn straight
// from an SVG, never by shrinking a bigger PNG.
//
// Needs @resvg/resvg-js, which is NOT a project dependency. Install it in a
// folder outside the repo, once:
//
//   mkdir <folder> && cd <folder> && npm init -y && npm install @resvg/resvg-js
//
// Then, from anywhere:
//
//   node build/icon/render.mjs [--resvg <folder>] [--sheet <file.png>] [--no-icns]
//
//   --resvg <folder>   where @resvg/resvg-js is installed. Also read from the
//                      MYCHAT_RESVG_DIR env var. Default:
//                      <temp dir>/opencode/icon-render
//   --sheet <file.png> also write a review sheet: every .ico size at 1x and
//                      at 8x nearest-neighbour on mid-grey
//   --sheet-sizes a,b  only these sizes on the sheet
//   --no-icns          skip build/darwin/icons.icns (which needs wails3)
//
// Writes:
//   build/appicon.png        1024x1024, from the main SVG
//   build/windows/icon.ico   16 20 24 32 40 48 64 96 128 256
//   build/trayicon.ico       16 20 24 32 40 48 64
//   build/darwin/icons.icns  from appicon.png, by `wails3 generate icons`
//
// Masters:
//   frontend/src/assets/mychat-icon.svg     48 px and up
//   build/icon/mychat-icon-small.svg        16 and 32 px (edges on a 64-unit grid)
//   build/icon/mychat-icon-small-20.svg     20 px, and 40 px at exactly 2x
//   build/icon/mychat-icon-small-24.svg     24 px

import { readFileSync, writeFileSync, mkdirSync } from 'node:fs'
import { createRequire } from 'node:module'
import { spawnSync } from 'node:child_process'
import { deflateSync } from 'node:zlib'
import { tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const buildDir = resolve(here, '..')
const repoDir = resolve(buildDir, '..')

const MAIN = join(repoDir, 'frontend', 'src', 'assets', 'mychat-icon.svg')
const SMALL = join(here, 'mychat-icon-small.svg')
const SMALL_20 = join(here, 'mychat-icon-small-20.svg')
const SMALL_24 = join(here, 'mychat-icon-small-24.svg')

// Which master draws which size.
const SOURCES = { 16: SMALL, 20: SMALL_20, 24: SMALL_24, 32: SMALL, 40: SMALL_20 }
const APP_SIZES = [16, 20, 24, 32, 40, 48, 64, 96, 128, 256]
const TRAY_SIZES = [16, 20, 24, 32, 40, 48, 64]

// ---- arguments -----------------------------------------------------------

const argv = process.argv.slice(2)
function option(name) {
  const i = argv.indexOf(name)
  if (i === -1) return undefined
  if (i + 1 >= argv.length) fail(`${name} needs a value`)
  return argv[i + 1]
}
function fail(message) {
  console.error('render.mjs: ' + message)
  process.exit(1)
}

const resvgDir = resolve(
  option('--resvg') ?? process.env.MYCHAT_RESVG_DIR ?? join(tmpdir(), 'opencode', 'icon-render'),
)
const sheetPath = option('--sheet')
const sheetSizes = option('--sheet-sizes')?.split(',').map(Number)
const skipIcns = argv.includes('--no-icns')

let Resvg
try {
  ;({ Resvg } = createRequire(join(resvgDir, 'noop.js'))('@resvg/resvg-js'))
} catch (err) {
  fail(
    `cannot load @resvg/resvg-js from ${resvgDir}\n` +
      `  install it there (npm install @resvg/resvg-js) or pass --resvg <folder>\n  ${err.message}`,
  )
}

// ---- rendering -----------------------------------------------------------

const svgCache = new Map()
function readSvg(path) {
  if (!svgCache.has(path)) svgCache.set(path, readFileSync(path, 'utf8'))
  return svgCache.get(path)
}

// One size, drawn from its master. Returns { size, png, pixels } where
// pixels is RGBA, 4 bytes per pixel.
const rendered = new Map()
function render(size) {
  if (!rendered.has(size)) {
    const svg = readSvg(SOURCES[size] ?? MAIN)
    const image = new Resvg(svg, { fitTo: { mode: 'width', value: size } }).render()
    if (image.width !== size || image.height !== size) {
      fail(`${size} px came out ${image.width}x${image.height}`)
    }
    rendered.set(size, { size, png: image.asPng(), pixels: Buffer.from(image.pixels) })
  }
  return rendered.get(size)
}

// ---- ICO -----------------------------------------------------------------

// ICONDIR (6 bytes), one ICONDIRENTRY (16 bytes) per image, then the PNG
// payloads. A width or height byte of 0 means 256.
function buildIco(sizes) {
  const images = [...sizes].sort((a, b) => a - b).map(render)
  const header = Buffer.alloc(6 + 16 * images.length)
  header.writeUInt16LE(0, 0) // reserved
  header.writeUInt16LE(1, 2) // type: icon
  header.writeUInt16LE(images.length, 4)
  let offset = header.length
  images.forEach(({ size, png }, i) => {
    const at = 6 + 16 * i
    header.writeUInt8(size === 256 ? 0 : size, at) // width
    header.writeUInt8(size === 256 ? 0 : size, at + 1) // height
    header.writeUInt8(0, at + 2) // palette colours
    header.writeUInt8(0, at + 3) // reserved
    header.writeUInt16LE(1, at + 4) // colour planes
    header.writeUInt16LE(32, at + 6) // bits per pixel
    header.writeUInt32LE(png.length, at + 8)
    header.writeUInt32LE(offset, at + 12)
    offset += png.length
  })
  return Buffer.concat([header, ...images.map((image) => image.png)])
}

const PNG_SIGNATURE = Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a])

// Reads the file back and checks the count, the order and every size, both
// in the directory and in each PNG's own header.
function validateIco(path, sizes) {
  const want = [...sizes].sort((a, b) => a - b)
  const file = readFileSync(path)
  if (file.readUInt16LE(0) !== 0 || file.readUInt16LE(2) !== 1) fail(`${path}: not an ICO`)
  const count = file.readUInt16LE(4)
  if (count !== want.length) fail(`${path}: ${count} images, expected ${want.length}`)
  let end = 6 + 16 * count
  want.forEach((size, i) => {
    const at = 6 + 16 * i
    const width = file.readUInt8(at) || 256
    const height = file.readUInt8(at + 1) || 256
    const bits = file.readUInt16LE(at + 6)
    const length = file.readUInt32LE(at + 8)
    const offset = file.readUInt32LE(at + 12)
    if (width !== size || height !== size) fail(`${path}: entry ${i} is ${width}x${height}, expected ${size}`)
    if (bits !== 32) fail(`${path}: entry ${i} is ${bits}-bit`)
    if (offset !== end || offset + length > file.length) fail(`${path}: entry ${i} has a bad offset`)
    const png = file.subarray(offset, offset + length)
    if (!png.subarray(0, 8).equals(PNG_SIGNATURE)) fail(`${path}: entry ${i} is not a PNG`)
    const pngWidth = png.readUInt32BE(16)
    const pngHeight = png.readUInt32BE(20)
    const bitDepth = png.readUInt8(24)
    const colourType = png.readUInt8(25)
    if (pngWidth !== size || pngHeight !== size) fail(`${path}: PNG ${i} is ${pngWidth}x${pngHeight}`)
    if (bitDepth !== 8 || colourType !== 6) fail(`${path}: PNG ${i} is not 8-bit RGBA`)
    end = offset + length
  })
  if (end !== file.length) fail(`${path}: ${file.length - end} stray bytes at the end`)
}

function writeIco(path, sizes) {
  mkdirSync(dirname(path), { recursive: true })
  writeFileSync(path, buildIco(sizes))
  validateIco(path, sizes)
  console.log(`${path}  ${[...sizes].sort((a, b) => a - b).join(' ')}  (checked)`)
}

// ---- review sheet --------------------------------------------------------

const CRC_TABLE = Array.from({ length: 256 }, (_, n) => {
  let c = n
  for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1
  return c >>> 0
})
function crc32(buffer) {
  let c = 0xffffffff
  for (const byte of buffer) c = CRC_TABLE[(c ^ byte) & 0xff] ^ (c >>> 8)
  return (c ^ 0xffffffff) >>> 0
}
function pngChunk(type, data) {
  const body = Buffer.concat([Buffer.from(type, 'ascii'), data])
  const out = Buffer.alloc(8 + data.length + 4)
  out.writeUInt32BE(data.length, 0)
  body.copy(out, 4)
  out.writeUInt32BE(crc32(body), 8 + data.length)
  return out
}
// Opaque RGB PNG from a width*height*3 buffer.
function encodePng(width, height, rgb) {
  const header = Buffer.alloc(13)
  header.writeUInt32BE(width, 0)
  header.writeUInt32BE(height, 4)
  header.writeUInt8(8, 8) // bit depth
  header.writeUInt8(2, 9) // colour type: RGB
  const raw = Buffer.alloc((width * 3 + 1) * height)
  for (let y = 0; y < height; y++) {
    rgb.copy(raw, y * (width * 3 + 1) + 1, y * width * 3, (y + 1) * width * 3)
  }
  return Buffer.concat([
    PNG_SIGNATURE,
    pngChunk('IHDR', header),
    pngChunk('IDAT', deflateSync(raw, { level: 9 })),
    pngChunk('IEND', Buffer.alloc(0)),
  ])
}

// 3x5 digits for the size labels.
const DIGITS = {
  0: '111101101101111', 1: '010110010010111', 2: '111001111100111', 3: '111001111001111',
  4: '101101111001001', 5: '111100111001111', 6: '111100111101111', 7: '111001001001001',
  8: '111101111101111', 9: '111101111001111',
}

function writeSheet(path, sizes) {
  const ZOOM = 8
  const PAD = 24
  const LABEL = 5 * 3 // digit height at 3x
  const MAX_ROW = 2200
  const GREY = 128

  // Each cell: label, the 8x image, then the 1x image under it.
  const cells = sizes.map((size) => ({
    size,
    width: size * ZOOM,
    height: LABEL + 8 + size * ZOOM + 12 + size,
  }))
  const rows = [[]]
  let rowWidth = PAD
  for (const cell of cells) {
    if (rows.at(-1).length && rowWidth + cell.width + PAD > MAX_ROW) {
      rows.push([])
      rowWidth = PAD
    }
    rows.at(-1).push(cell)
    rowWidth += cell.width + PAD
  }
  const width = Math.max(...rows.map((row) => PAD + row.reduce((sum, c) => sum + c.width + PAD, 0)))
  const height = PAD + rows.reduce((sum, row) => sum + Math.max(...row.map((c) => c.height)) + PAD, 0)
  const rgb = Buffer.alloc(width * height * 3, GREY)

  const put = (x, y, r, g, b, a = 255) => {
    const at = (y * width + x) * 3
    rgb[at] = Math.round((r * a + rgb[at] * (255 - a)) / 255)
    rgb[at + 1] = Math.round((g * a + rgb[at + 1] * (255 - a)) / 255)
    rgb[at + 2] = Math.round((b * a + rgb[at + 2] * (255 - a)) / 255)
  }
  const blit = ({ size, pixels }, left, top, zoom) => {
    for (let y = 0; y < size; y++) {
      for (let x = 0; x < size; x++) {
        const at = (y * size + x) * 4
        for (let dy = 0; dy < zoom; dy++) {
          for (let dx = 0; dx < zoom; dx++) {
            put(left + x * zoom + dx, top + y * zoom + dy, pixels[at], pixels[at + 1], pixels[at + 2], pixels[at + 3])
          }
        }
      }
    }
  }
  const label = (text, left, top) => {
    ;[...text].forEach((digit, i) => {
      const bits = DIGITS[digit]
      for (let n = 0; n < 15; n++) {
        if (bits[n] !== '1') continue
        for (let dy = 0; dy < 3; dy++) {
          for (let dx = 0; dx < 3; dx++) put(left + i * 12 + (n % 3) * 3 + dx, top + Math.floor(n / 3) * 3 + dy, 255, 255, 255)
        }
      }
    })
  }

  let top = PAD
  for (const row of rows) {
    let left = PAD
    for (const cell of row) {
      const image = render(cell.size)
      label(String(cell.size), left, top)
      blit(image, left, top + LABEL + 8, ZOOM)
      blit(image, left, top + LABEL + 8 + cell.size * ZOOM + 12, 1)
      left += cell.width + PAD
    }
    top += Math.max(...row.map((c) => c.height)) + PAD
  }

  mkdirSync(dirname(path), { recursive: true })
  writeFileSync(path, encodePng(width, height, rgb))
  console.log(`${path}  ${width}x${height}`)
}

// ---- run -----------------------------------------------------------------

const appIcon = join(buildDir, 'appicon.png')
const appImage = new Resvg(readSvg(MAIN), { fitTo: { mode: 'width', value: 1024 } }).render()
writeFileSync(appIcon, appImage.asPng())
console.log(`${appIcon}  ${appImage.width}x${appImage.height}`)

writeIco(join(buildDir, 'windows', 'icon.ico'), APP_SIZES)
writeIco(join(buildDir, 'trayicon.ico'), TRAY_SIZES)

if (sheetPath) writeSheet(resolve(sheetPath), sheetSizes ?? APP_SIZES)

if (!skipIcns) {
  // wails3 always writes a Windows .ico too; send it to the temp folder so
  // the one made above is not overwritten.
  const throwaway = join(tmpdir(), 'mychat-wails-icon.ico')
  const result = spawnSync(
    'wails3',
    ['generate', 'icons', '-input', 'appicon.png', '-macfilename', 'darwin/icons.icns', '-windowsfilename', throwaway],
    { cwd: buildDir, stdio: 'inherit' },
  )
  if (result.error || result.status !== 0) {
    fail(`wails3 generate icons failed${result.error ? ': ' + result.error.message : ''}`)
  }
  console.log(join(buildDir, 'darwin', 'icons.icns'))
}
