/** @type {import('tailwindcss').Config} */
export default {
    content: [
        "./index.html",
        "./src/**/*.{vue,js,ts,jsx,tsx}",
    ],
    darkMode: 'class',
    theme: {
        extend: {
            colors: {
                primary: '#0066ff',
                bgApp: 'var(--bg-app)',
                bgCard: 'var(--bg-card)',
                textMain: 'var(--text-main)',
                textLight: 'var(--text-light)',
            },
            borderRadius: {
                lg: '8px',
                xl: '12px',
                '2xl': '16px',
                '3xl': '24px',
            },
            boxShadow: {
                card: '0 4px 12px rgba(0,0,0,0.05)',
                premium: '0 10px 30px -10px rgba(0,0,0,0.1)',
            },
        },
    },
    plugins: [],
}
