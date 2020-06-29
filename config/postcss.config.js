module.exports = ctx => {
  let config = {
    map: ctx.options.map,
    parser: ctx.options.parser,
    plugins: {
      "autoprefixer": {},
    },
  };

  if (ctx.env === 'production') {
    Object.assign(config.plugins, {
      "cssnano": {},
    });
  }

  return config;
};
