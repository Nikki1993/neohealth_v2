const defaultTheme = require('tailwindcss/defaultTheme')

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./templates/**/*.gohtml'],
  theme: {
    extend: {
      fontFamily: {
        sans:  ['Inter', ...defaultTheme.fontFamily.sans],
        title: ['Caveat']
      }
    },
  },
  plugins: [],
}
