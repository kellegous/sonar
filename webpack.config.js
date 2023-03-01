const path = require('path'),
	dist = path.resolve(__dirname, 'pkg/web/ui');

module.exports = {
	entry: {
		'index': './ui/index.ts',
	},
	output: {
		filename: "[name].js",
		path: dist,
		clean: false,
	},
	resolve: {
		extensions: ['.ts', '.tsx', '.js'],
	},
	module: {
		rules: [
			{
				test: /\.tsx?$/,
				use: "ts-loader"
			},
			{
				test: /\.s?css$/,
				use: [
					"style-loader",
					"css-loader",
					"sass-loader",
				]
			}
		]
	},
	plugins: [],
	devServer: {
		static: dist
	},
};
