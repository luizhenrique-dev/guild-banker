import tseslint from 'typescript-eslint'
import hooks from 'eslint-plugin-react-hooks'

export default tseslint.config(
  { ignores: ['node_modules/**', 'dist/**', 'public/mockServiceWorker.js', 'src/shared/api/schema.d.ts', 'playwright-report/**', 'test-results/**', 'scripts/safe-seed.ts'] },
  ...tseslint.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    plugins: { 'react-hooks': hooks },
    rules: { ...hooks.configs.recommended.rules, '@typescript-eslint/no-unused-vars': ['error', { varsIgnorePattern: '^actionTypes$' }] },
  },
)
