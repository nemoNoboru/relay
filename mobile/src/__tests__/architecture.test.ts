import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative, sep } from 'node:path'

// Dependency rule: domain <- application <- infrastructure / presentation; only composition sees everything.
const layers = ['domain', 'application', 'infrastructure', 'presentation', 'composition'] as const
type Layer = (typeof layers)[number]

const allowedLayers: Record<Layer, readonly Layer[]> = {
  domain: ['domain'],
  application: ['domain', 'application'],
  infrastructure: ['domain', 'infrastructure'],
  presentation: ['domain', 'application', 'presentation'],
  composition: layers,
}

// Third-party packages a layer must not touch.
const forbiddenPackages: Partial<Record<Layer, RegExp>> = {
  domain: /^(react|react-native|axios|expo|@react-native|jest)/,
  application: /^(react|react-native|axios|expo|@react-native)/,
}

const srcDir = join(__dirname, '..')

function sourceFiles(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const path = join(dir, name)
    if (statSync(path).isDirectory()) return name === '__tests__' ? [] : sourceFiles(path)
    return /\.tsx?$/.test(name) ? [path] : []
  })
}

const importsOf = (file: string): string[] =>
  [...readFileSync(file, 'utf8').matchAll(/(?:from|import)\s+['"]([^'"]+)['"]/g)].map((m) => m[1]!)

const layerOf = (file: string): Layer | undefined => layers.find((l) => relative(srcDir, file).startsWith(l + sep))

describe.each(layers)('%s layer', (layer) => {
  const files = sourceFiles(join(srcDir, layer))

  it('has source files', () => {
    expect(files.length).toBeGreaterThan(0)
  })

  it.each(files.map((f) => [relative(srcDir, f), f]))('%s respects the dependency rule', (_name, file) => {
    for (const spec of importsOf(file!)) {
      if (spec.startsWith('.')) {
        const target = layerOf(join(file!, '..', spec))
        if (target) expect({ spec, allowed: allowedLayers[layer].includes(target) }).toEqual({ spec, allowed: true })
      } else {
        expect({ spec, forbidden: forbiddenPackages[layer]?.test(spec) ?? false }).toEqual({ spec, forbidden: false })
      }
    }
  })
})
