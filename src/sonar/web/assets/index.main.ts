/// <reference path="lib/xhr.ts" />
/// <reference path="lib/signal.ts" />
/// <reference path="lib/dom.ts" />

module app {

	const SVGNS = 'http://www.w3.org/2000/svg';

	interface Res<T> {
		ok: boolean;
		data: T;
	}

	interface Summary {
		p10: number;
		p50: number;
		p90: number;
		count: number;
		max: number;
		min: number;
		lossRatio: number;
		data?: number[];
	}

	interface Current extends Summary {
		ip: string;
		name: string;
		time: string;
	}

	interface Hour extends Summary {
		time: string;
	}

	interface Hourly {
		ip: string;
		name: string;
		hours: Hour[];
	}

	function formatNumber(num: number, dec: number) : string {
		var w = num | 0,
			ws = '' + w,
			f = num - w,
			res = [];
		while (ws.length > 0) {
			res.unshift(ws.substring(ws.length - 3, ws.length));
			ws = ws.substring(0, ws.length - 3);
		}

		if (dec < 1) {
			return res.join(',');
		}

		var fs = f.toFixed(dec);
		return res.join(',') + fs.substring(1);
	}

	function renderMs(el: HTMLElement, ns: number) {
		var ms = ns / 1e6;
		el.textContent = formatNumber(ms, (ms > 1000) ? 0 : 1);

		var uel = document.createElement('span');
		uel.classList.add('unit-ms');
		uel.textContent = 'ms';
		el.appendChild(uel);
	}

	function formatMs(ns: number, dec: number) : string {
		var ms = ns / 1e6;
		return formatNumber(ms, dec) + 'ms';
	}

	function formatLossPercent(el: HTMLElement, p: number) {
		if (p < 0.00001) {
			el.textContent = '0';
		} else if (p > 0.99999) {
			el.textContent = '100';
		} else {
			el.textContent = (p*100).toFixed(1);
		}

		var uel = document.createElement('span');
		uel.classList.add('unit-perc');
		uel.textContent = '%';
		el.appendChild(uel);
	}

	function formatHour(dt: Date) : string {
		var h = '' + dt.getHours();
		return h.length == 1 ? '0' + h : h;
	}

	interface Report {
		ip: string;
		name: string;
		currently: Summary;
		hourly: Hour[];
	}

	function toValues<T>(obj: {[key: string]: T}): T[] {
		var vals: T[] = [];
		for (var key in obj) {
			vals.push(obj[key]);
		}
		return vals;
	};

	function start(period: number) : Q.Signal {
		var s = new Q.Signal();

		var load = () => {
			k4.mget('/api/v1/current', '/api/v1/hourly?n=48')
				.onSuccess((json: string[]) => {
					var curr = <Res<Current[]>>JSON.parse(json[0]),
						hrly = <Res<Hourly[]>>JSON.parse(json[1]),
						data: {[key: string]: Report} = {};

					curr.data.forEach((c) => {
						data[c.ip] = {
							ip: c.ip,
							name: c.name,
							currently: c,
							hourly: []
						};
					});

					hrly.data.forEach((h) => {
						data[h.ip].hourly = h.hours;
					});

					s.raise(toValues(data));
				});
		};

		setInterval(load, period);
		load();

		return s;
	}

	function renderRow(
		el: HTMLElement,
		fn: (root: dom.El, l: dom.El, r: dom.El) => void) {
		var root = dom.of(document.createElement('div'))
			.addClass('row')
			.appendTo(el);

		var l = dom.of(document.createElement('div'))
			.addClass('l')
			.appendTo(root.el);

		var r = dom.of(document.createElement('div'))
			.addClass('r')
			.appendTo(root.el);

		fn(root, l, r);
	}

	interface Range {
		min: number;
		max: number;
	}

	interface Scale {
		rng: Range;
		divs: number[];
		step: number;
	}

	function rangeFrom(hours: Hour[]) : Range {
		return hours.reduce((r: Range, hour: Hour) => {
			var max = hour.p90,
				min = hour.p10;

			if (hour.count == 0) {
				return r;
			}

			if (max > r.max) {
				r.max = max;
			}

			if (r.min == 0 || min < r.min) {
				r.min = min;
			}

			return r;
		}, {min: 0, max: 0});
	}

	function log10(x: number) : number {
		return Math['log10'](x);
	}

	function scaleFor(rng: Range) : Scale {
		var dy = rng.max - rng.min,
			mag = Math.pow(10, Math.floor(log10(dy) - 1)),
			facs = [1, 2, 5],
			lim = 3.5,
			expand = (s: number) => {
				var r = [],
					b = ((rng.min / s)|0)*s;
				for (var i = b + s; i < rng.max; i += s) {
					r.push(i);
				}
				return r;
			};

		while (true) {
			for (var i = 0, n = facs.length; i < n; i++) {
				var s = facs[i] * mag;
				if ((dy / s) <= lim) {
					return {
						rng: rng,
						divs: expand(s),
						step: s,
					};
				}
			}
			mag *= 10;
		}
	}

	function renderLossGraph(el: HTMLElement, report: Report) {
		var rect = el.getBoundingClientRect(),
			w = rect.width,
			h = rect.height,
			pad = 35,
			lim = 0.25,
			tpad = 20,
			bpad = 2,
			dx = (w - pad) / report.hourly.length;

		var svg = dom.create('svg', SVGNS)
			.setAttrs({
				width: w + 'px',
				height: h + 'px'
			})
			.appendTo(el)
			.rel();

		dom.create('line', SVGNS)
			.setAttrs({
				x1: 0,
				y1: tpad - bpad,
				x2: w,
				y2: tpad - bpad,
				stroke: '#eee',
				'stroke-dasharray': '1,4',
			})
			.appendTo(svg);

		dom.create('text', SVGNS)
			.setAttrs({
				x: 0,
				y: tpad - bpad + 10,
				fill: '#fff',
				'font-family': 'Roboto',
				'font-size': 9,
			})
			.setText((lim*100 | 0) + '%')
			.appendTo(svg);

		report.hourly.forEach((hr: Hour, i: number) => {
			if (hr.lossRatio < 0.001) {
				return;
			}

			var v = Math.min(lim, hr.lossRatio)/lim;
			dom.create('rect', SVGNS)
				.setAttrs({
					x: pad + dx*i + 3,
					y: h - (h-tpad)*v - bpad,
					width: dx - 6,
					height: (h-tpad)*v,
					fill: '#eee',
				})
				.appendTo(svg);

			var p = Math.min(99, (hr.lossRatio*100)|0);
			if (p < 1) {
				return;
			}

			dom.create('text', SVGNS)
				.setAttrs({
					x: pad + dx*i + (p > 9 ? 1 : 4),
					y: h - (h-tpad)*v - 5,
					fill: '#fff',
					'font-family': 'Roboto',
					'font-size': 8,
				})
				.setText(p + '%')
				.appendTo(svg);
		});
	}

	function renderTimeGraph(el: HTMLElement, report: Report) {
		var rect = el.getBoundingClientRect(),
			w = rect.width,
			h = rect.height - 20,
			scale = scaleFor(rangeFrom(report.hourly)),
			min = scale.rng.min,
			max = scale.rng.max,
			log = Math.floor(log10(scale.step)),
			pad = 35,
			dx = (w - pad) / report.hourly.length,
			dy = h / (max - min);

		var svg = dom.of(document.createElementNS(SVGNS, 'svg'))
			.setAttrs({
				width: w + 'px',
				height: rect.height + 'px',
			})
			.appendTo(el)
			.rel();

		scale.divs.forEach((div) => {
			var y = h - dy*div + min*dy;
			dom.create('line', SVGNS)
				.setAttrs({
					x1: 0,
					y1: y,
					x2: w,
					y2: y,
					stroke: '#eee',
					'stroke-dasharray': '1,4',
				})
				.appendTo(svg);

			// if the text is likely to get clipped, don't show it.
			if (y < 15) {
				return;
			}

			dom.create('text', SVGNS)
				.setAttrs({
					x: 0,
					y: y - 3,
					fill: '#fff',
					'font-family': 'Roboto',
					'font-size': 9,
				})
				.setText(formatMs(div, log < 0 ? -log : 0))
				.appendTo(svg);
		});

		report.hourly.forEach((hr: Hour, i: number) => {
			var t = new Date(Date.parse(hr.time));
			dom.create('text', SVGNS)
				.setAttrs({
					x: pad + dx*i + 3,
					y: h + 15,
					fill: '#fff',
					'font-family': 'Roboto',
					'font-size': 9,
				})
				.setText(formatHour(t))
				.appendTo(svg);

			if (hr.count == 0) {
				return;
			}
			
			dom.create('rect', SVGNS)
				.setAttrs({
					x: pad + dx*i + 3,
					y: h - dy*hr.p90 + min*dy,
					width: dx - 6,
					height: (hr.p90 - hr.p10)*dy,
					// fill: '#f19fa7',
					fill: '#a41742',
					// fill: 'rgba(164,23,66,0.6)',
				})
				.appendTo(svg);

			dom.create('rect', SVGNS)
				.setAttrs({
					x: pad + dx*i + 2,
					y: h - dy*hr.p50+ min*dy,
					width: dx - 4,
					height: 2,
					fill: '#fff',
				})
				.appendTo(svg);
		});
	}

	function render(el: HTMLElement, reports: Report[]) {
		el.textContent = '';
		reports.forEach((report) => {
			dom.of(document.createElement('div'))
				.addClass('host')
				.appendTo(el)
				.do((host) => {
					var head = dom.of(document.createElement('div'))
						.addClass('head')
						.appendTo(host.el)
						.rel();

					dom.of(document.createElement('span'))
						.addClass('name')
						.setText(report.name)
						.appendTo(head);

					dom.of(document.createElement('span'))
						.addClass('ip')
						.setText(report.ip)
						.appendTo(head);

					renderRow(host.el, (root, l, r) => {
						root.addClass('time');
						l.addClass('curr');

						renderMs(l.el, report.currently.p50);
						dom.create('div')
							.addClass('label')
							.setText('rtt')
							.appendTo(l.el);

						r.addClass('graf');
						renderTimeGraph(r.el, report);
					});

					renderRow(host.el, (root, l, r) => {
						root.addClass('loss');

						l.addClass('curr');
						formatLossPercent(l.el, report.currently.lossRatio);
						dom.create('div')
							.addClass('label')
							.setText('loss')
							.appendTo(l.el);

						r.addClass('graf');
						renderLossGraph(r.el, report);
					});
				});
		});
	};

	start(60*1000).tap((reports: Report[]) => {
		render(<HTMLElement>document.querySelector('#cnt'), reports);
	});
}