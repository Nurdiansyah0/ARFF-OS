import forms from '@tailwindcss/forms';

/** @type {import('tailwindcss').Config} */
export default {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      fontFamily: {
        sans: ['system-ui', '-apple-system', 'BlinkMacSystemFont', 'Segoe UI', 'Roboto', 'sans-serif'],
      },
      colors: {
        primary: {
          red: '#C62828',
          redHover: '#B71C1C',
          gold: '#E3A008',
          goldHover: '#CA8A04',
        },
        semantic: {
          success: '#16A34A',
          warning: '#EAB308',
          danger: '#DC2626',
          info: '#2563EB',
          neutral: '#64748B',
        },
        brand: '#C62828',
      },
    },
  },
  plugins: [forms],
};
