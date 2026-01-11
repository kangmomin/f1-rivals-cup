/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Brand Colors
        carbon: {
          DEFAULT: '#121212',
          light: '#1E1E1E',
          dark: '#0A0A0A',
        },
        neon: {
          DEFAULT: '#0A84FF',
          light: '#409CFF',
          dark: '#0066CC',
        },
        racing: {
          DEFAULT: '#FF3B30',
          light: '#FF6961',
          dark: '#CC2F26',
        },
        // Semantic Colors
        steel: {
          DEFAULT: '#3A3A3C',
          light: '#505052',
          dark: '#2C2C2E',
        },
        profit: '#30D158',
        loss: '#FF453A',
        warning: '#FF9F0A',
        // Text Colors
        'text-primary': '#FFFFFF',
        'text-secondary': '#A0A0A0',
        'text-muted': '#6C6C6C',
      },
      fontFamily: {
        heading: ['Saira', 'Rajdhani', 'sans-serif'],
        body: ['Inter', 'Pretendard', 'sans-serif'],
        mono: ['JetBrains Mono', 'Roboto Mono', 'monospace'],
      },
      spacing: {
        xs: '4px',
        sm: '8px',
        md: '16px',
        lg: '24px',
        xl: '32px',
      },
      borderRadius: {
        sm: '4px',
        md: '8px',
        lg: '12px',
      },
      transitionDuration: {
        fast: '150ms',
        normal: '300ms',
      },
    },
  },
  plugins: [],
}
