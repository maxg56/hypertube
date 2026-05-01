'use client';

import React from 'react';
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import enTranslation from '@/locales/en.json';
import frTranslation from '@/locales/fr.json';

const resources = {
  en: {
    translation: enTranslation,
  },
  fr: {
    translation: frTranslation,
  },
};

if (!i18n.isInitialized) {
  i18n
    .use(initReactI18next)
    .init({
      resources,
      lng: 'fr',
      fallbackLng: 'en',
      interpolation: {
        escapeValue: false,
      },
    });
}

export function I18nProvider({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}
