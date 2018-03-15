package v1

const Console_JS = `
web3._extend({
	property: 'cht',
	methods: [
		new web3._extend.Method({
			name: 'post',
			call: 'cht_post',
			params: 1
		}),
		new web3._extend.Method({
			name: 'poll',
			call: 'cht_poll',
			params: 1
		}),
		new web3._extend.Method({
			name: 'pollStr',
			call: 'cht_pollStr',
			params: 1
		}),
	],
	properties:
	[   new web3._extend.Property({
			name: 'version',
			getter: 'cht_version',
			outputFormatter: web3._extend.utils.toDecimal
		})
	],
});
var cht = web3.cht;
`
