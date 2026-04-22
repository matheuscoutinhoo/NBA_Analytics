const nextJest = require('next/jest');

const createJestConfig = nextJest({
   dir: './',
});

const customJestConfig = {
   setupFilesAfterEnv: ['<rootDir>/jest.setup.ts'],
   testEnvironment: 'jsdom',
   moduleNameMapper: {
      '^@/(.*)$': '<rootDir>/$1',
   },
   testPathIgnorePatterns: ['<rootDir>/.next/', '<rootDir>/node_modules/'],
   coverageThreshold: {
      global: {
         branches: 50,
         functions: 50,
         lines: 70,
         statements: 70,
      },
   },
};

module.exports = createJestConfig(customJestConfig);
