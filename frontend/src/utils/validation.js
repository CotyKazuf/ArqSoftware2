export const isValidEmail = (value = '') =>
  /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value.trim().toLowerCase())

export const hasMinLength = (value = '', min = 2) => value.trim().length >= min

export const normalizeText = (value = '') =>
  value
    .toString()
    .toLowerCase()
    .normalize('NFD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/[^a-z0-9 ]/g, '')
    .trim()
