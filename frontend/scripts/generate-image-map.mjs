import { readFileSync, readdirSync, writeFileSync, statSync } from 'node:fs'
import { join, sep } from 'node:path'

const ROOT = process.cwd()
const dataPath = join(ROOT, 'src', 'data', 'perfumes.json')

const normalize = (value) =>
  value
    ?.toString()
    .toLowerCase()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/&/g, 'and')
    .replace(/[^a-z0-9]/g, '') ?? ''

const fileMap = new Map()
const manualOverrides = new Map([
  ['parfumsdemarlydelina', 'img/perfumes/Delina (Parfums de Marly).jpg'],
  ['azzaroazzarothemostwantedintense', 'img/perfumes/Azzaro Most Wanted Intense EDP.jpg'],
  ['viktorrolfspicebomb', 'img/perfumes/Spicebomb Viktor&amp.webp'],
])

const sources = [
  { dir: join(ROOT, 'public', 'img', 'perfumes'), publicPrefix: 'img/perfumes' },
  { dir: join(ROOT, 'public', 'Imagenes_perfumes', 'Imagenes_perfumes_mujer'), publicPrefix: 'Imagenes_perfumes/Imagenes_perfumes_mujer' },
  { dir: join(ROOT, 'public', 'Imagenes_perfumes', 'Imagenes_perfumes_hombre'), publicPrefix: 'Imagenes_perfumes/Imagenes_perfumes_hombre' },
]

function collectFiles(dir, publicPrefix) {
  let entries
  try {
    entries = readdirSync(dir, { withFileTypes: true })
  } catch {
    return
  }

  for (const entry of entries) {
    const entryPath = join(dir, entry.name)
    if (entry.isDirectory()) {
      collectFiles(entryPath, `${publicPrefix}/${entry.name}`)
      continue
    }

    const baseName = entry.name.replace(/\.[^.]+$/, '')
    const key = normalize(baseName)
    if (!key || fileMap.has(key)) continue
    const relPath = `${publicPrefix}/${entry.name}`.split(sep).join('/')
    fileMap.set(key, relPath)
  }
}

for (const source of sources) {
  if (!statExists(source.dir)) continue
  collectFiles(source.dir, source.publicPrefix)
}

function statExists(path) {
  try {
    return statSync(path).isDirectory()
  } catch {
    return false
  }
}

const perfumes = JSON.parse(readFileSync(dataPath, 'utf8'))
const missing = []

for (const perfume of perfumes) {
  const slugKey = normalize(perfume.imagen?.split('/').pop()?.replace(/\.[^.]+$/, ''))
  const nameKey = normalize(perfume.nombre)
  const brandNameKey = normalize(`${perfume.marca ?? ''} ${perfume.nombre ?? ''}`)
  const nameBrandKey = normalize(`${perfume.nombre ?? ''} ${perfume.marca ?? ''}`)
  const idKey = normalize(perfume.id)

  const keys = [slugKey, brandNameKey, nameBrandKey, nameKey, idKey].filter(Boolean)
  const match =
    keys.map((key) => fileMap.get(key)).find(Boolean) ??
    keys.map((key) => manualOverrides.get(key)).find(Boolean)

  if (match) {
    perfume.imagen = match
  } else {
    missing.push({ id: perfume.id, nombre: perfume.nombre })
    perfume.imagen = 'img/principal.jpg'
  }
}

writeFileSync(dataPath, `${JSON.stringify(perfumes, null, 2)}\n`, 'utf8')

if (missing.length) {
  console.warn('No se encontraron imágenes para:', missing)
} else {
  console.log('Todas las fragancias quedaron asociadas a una imagen.')
}
