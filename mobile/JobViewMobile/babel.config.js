module.exports = {
  presets: ['module:@react-native/babel-preset'],
  plugins: [
    [
      'module-resolver',
      {
        root: ['./src'],
        alias: {
          '@': './src',
          '@components': './src/components',
          '@screens': './src/screens',
          '@services': './src/services',
          '@store': './src/store',
          '@utils': './src/utils',
          '@hooks': './src/hooks',
          '@types': './src/types',
          '@navigation': './src/navigation',
          '@config': './src/config',
          '@constants': './src/constants',
          '@assets': './src/assets',
        },
      },
    ],
  ],
};
