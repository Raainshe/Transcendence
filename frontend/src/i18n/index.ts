import { createI18n } from 'vue-i18n'

import de from '@/locales/de.json'
import en from '@/locales/en.json'
import fr from '@/locales/fr.json'

export const SUPPORTED_LOCALES = ['en', 'de', 'fr'] as const
export type AppLocale = (typeof SUPPORTED_LOCALES)[number]

export const DEFAULT_LOCALE: AppLocale = 'en'

export const i18n = createI18n({
  legacy: false,
  locale: DEFAULT_LOCALE,
  fallbackLocale: DEFAULT_LOCALE,
  messages: {
    en,
    de,
    fr,
  },
})
