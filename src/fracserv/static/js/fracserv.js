String.prototype.capitalize = function() {
	return this.charAt(0).toUpperCase() + this.slice(1);
}

var map;

!function( $ ) {

	$(function () {

		var setMapLocation = function () {
			var fracParams = location.hash.split('?')[1];
			console.log("fracParams", fracParams);
			var urlParams = getUrlVars('?' + fracParams);
			if('z' in urlParams) {
				map.setZoom(parseInt(urlParams.z));
			} else {
				console.log("Setting zoom to 0");
				map.setZoom(0);
			}

			if('c' in urlParams) {
				var v = urlParams.c.split(',');
				console.log("v", v);
				var lat = parseFloat(v[0]),
					lng = parseFloat(v[1]);

				var center = new google.maps.LatLng(lat, lng, true);
				map.setCenter(center);
				map.panTo(center);
			} else {
				console.log("Centering on 0");
				var center = new google.maps.LatLng(0, 0);
				map.setCenter(center);
			}
		};

		var getContents = function(fracType) {
			var fracParams = location.hash.split('?')[1];
			urlParams = getUrlVars('?' + fracParams);

			$('#config-content').load('/' + fracType + ' form', function() {
				var form = $('#config-content form');
				map = initialize(fracType);
				var emptyForm = $('input', form).length == 0;
				$('#config').toggle(!emptyForm);
				if(!emptyForm) {
					$('input', form).bind('input', function() {
						console.log('Form changed, redrawing');
						map.reload();
					});

					form.submit(function() {
						console.log('Form submitted');
						map.reload();
						return false;
					});
				}

				var resize = function() {
					var navHeight = $('.navbar-fixed-top').height();
					$('#maps').width($(window).width())
							  .height($(window).height()-navHeight)
							  .css('top', navHeight+'px');
					google.maps.event.trigger(map, 'resize')
				}

				// Populate form from parameters
				$('form input').each(function(idx, e) {
					if(e.id in urlParams) {
						e.value = urlParams[e.id];
					}
				});

				setMapLocation();
				$(window).resize(function() {
					resize();
				});
				$('a.btn.btn-navbar').click(resize);
				resize();
				location.hash = '#' + fracType;
			});
		};
		var dismiss = function() {
			$('#config').fadeOut();
			$('#gear').fadeIn();
		};

		var show = function() {
			$('#config').fadeIn();
			$('#gear').fadeOut();
		};
		$('#hide').click(dismiss);
		$('#share').click(function() {
			var url = location.origin + '/' + map.fracSave()
			$('#share-modal #share-url a')
				.html(url)
				.attr('href', url);

			$('#share-modal').modal('show');
		});
		$('#gear').click(show);

		$('ul.nav li a').click(function(e) {
			var fracType = $(this).attr('id');
			console.log(fracType);
			getContents(fracType);
			$('#masthead').fadeOut();
			$('#mobile-jump').fadeOut();
			$('#config').fadeIn();
		});

		var fracLoad = function() {
			var fracType = location.hash.split('?')[0];
			$(fracType).click();
		};
		fracLoad();
	})
}( window.jQuery );
