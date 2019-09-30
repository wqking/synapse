let mix = require('laravel-mix');

const publicFolder = 'public/';

mix.js('js/app.js', publicFolder).sass('sass/app.scss', publicFolder);

