(function () {

    class ListPage extends ListingPage {
        constructor() {
            super();
            this.totalPrice = document.getElementById('total-price');
            this.sortButton = document.getElementById('sort-button');
            this.sortButton.addEventListener('click', () => {
                const hideLoader = showOverlayLoader();
                this.sort().catch(handleError).finally(hideLoader);
            });
            this.routeButton = document.getElementById('route-button');
            this.routeButton.addEventListener('click', () => {
                window.open('/route?store=' + encodeURIComponent(currentStore()));
            });
        }

        async sort() {
            await this.waitForInitialData();
            const response = await fetch('/api/sort?store=' + encodeURIComponent(currentStore()), {
                credentials: 'same-origin',
                cache: 'no-store',
            });
            const data = await response.json();
            if (data.error) {
                throw data.error;
            }
            this.updateData(data);
        }

        async fetchData() {
            const response = await fetch('/api/list?store=' + encodeURIComponent(currentStore()), {
                credentials: 'same-origin',
                cache: 'no-store',
            });
            const data = await response.json();
            if (data.error) {
                throw data.error;
            }
            return data;
        }

        showList() {
            super.showList();
            this.totalPrice.style.display = 'block';
            this.sortButton.style.display = 'inline-block';
            this.routeButton.style.display = 'inline-block';
        }

        hideList() {
            super.hideList();
            this.totalPrice.style.display = 'none';
            this.sortButton.style.display = 'none';
            this.routeButton.style.display = 'none';
        }

        dataChanged() {
            let totalPrice = 0;
            let notCorrect = false;
            this.data.forEach((item) => {
                // Price strings are either dollar amounts (e.g. "$150.00") or
                // ranges (e.g. "$99.99 - $149.99").
                // In the later case, we use the upper bound.
                const price = parseFloat(item.price.split(' ').pop().substr(1));
                if (!isNaN(price)) {
                    totalPrice += price;
                } else {
                    notCorrect = true;
                }
            })
            this.totalPrice.textContent = '$' + totalPrice.toFixed(2);
            if (notCorrect) {
                this.totalPrice.textContent += ' (missing items)';
            }
        }

        createAddDialog() {
            return new AddProductDialog();
        }

        createListItem(item) {
            return createListItem(item);
        }

        selectedListItem(item) {
            showProductInfo(item);
        }

        async addItem(item) {
            await this.addItemWithZone(item, null);
        }

        async addItemWithZone(item, zoneName) {
            let formData = 'store=' + encodeURIComponent(currentStore()) +
                '&signature=' + encodeURIComponent(item.signature) +
                '&data=' + encodeURIComponent(JSON.stringify(item.data));
            if (zoneName !== null) {
                formData += '&zone=' + encodeURIComponent(zoneName);
            }
            const response = await fetch('/api/additem', {
                method: 'POST',
                credentials: 'same-origin',
                headers: {
                    'content-type': 'application/x-www-form-urlencoded',
                },
                body: formData,
                cache: 'no-store',
            });
            const data = await response.json();
            if (data.noZone && zoneName === null) {
                await this.addDialog.locationPicker.open(data.error, (zoneName) => {
                    const hideLoader = showOverlayLoader();
                    this.addItemWithZone(item, zoneName)
                        .catch(handleError)
                        .finally(hideLoader);
                });
                return;
            }
            if (data.error) {
                throw data.error;
            }
            this.updateData(data);
            this.addDialog.close();
        }

        async deleteItem(item) {
            const query = '?store=' + encodeURIComponent(currentStore()) +
                '&item=' + encodeURIComponent(item.id);
            const response = await fetch('/api/removeitem' + query, {
                credentials: 'same-origin',
                cache: 'no-store',
            });
            const data = await response.json();
            if (data.error) {
                throw data.error;
            }
            this.updateData(data);
        }
    }

    class AddProductDialog extends AddDialog {
        constructor() {
            super();
            this.locationPicker = new LocationPicker();
        }

        async fetchSearchResults(query) {
            const queryStr = '?store=' + encodeURIComponent(currentStore()) +
                '&query=' + encodeURIComponent(query);
            const response = await fetch('/api/inventoryquery' + queryStr, {
                credentials: 'same-origin',
                cache: 'no-store',
            });
            const result = await response.json();
            if (result.error) {
                throw result.error;
            }
            return result.results.filter((x) => x.inStock);
        }

        createListItem(item) {
            return createListItem(item);
        }

        reset() {
            super.reset();
            this.locationPicker.close();
        }
    }

    class LocationPicker {
        constructor() {
            this.element = document.getElementById('location-picker');
            this.errorMessage = document.getElementById('location-picker-error');
            this.cancelButton = document.getElementById('location-picker-cancel');
            this.mapContainer = document.getElementById('location-picker-map');
            this.textLabels = [];

            this.cancelButton.addEventListener('click', () => this.close());
        }

        async open(errorMessage, onChosen) {
            const queryStr = '?store=' + encodeURIComponent(currentStore());
            const response = await fetch('/api/map' + queryStr, {
                credentials: 'same-origin',
                cache: 'no-store',
            });
            const svgData = await response.text();
            this.errorMessage.textContent = errorMessage;
            this.mapContainer.innerHTML = svgData;
            this.element.style.display = 'block';

            this.findTextElements();
            this.registerMouseEvents(onChosen);
        }

        close() {
            this.element.style.display = 'none';
        }

        findTextElements() {
            const svgElement = this.mapContainer.getElementsByTagName('svg')[0];
            const texts = svgElement.getElementsByTagName('text');
            const counts = {};
            for (let i = 0; i < texts.length; ++i) {
                const text = texts[i].textContent;
                counts[text] = (counts[text] || 0) + 1;
            }
            this.textLabels = [];
            for (let i = 0; i < texts.length; ++i) {
                const text = texts[i].textContent;
                // Only find unique zone names.
                if (counts[text] === 1) {
                    this.textLabels.push(texts[i]);
                }
            }
        }

        registerMouseEvents(onChosen) {
            const svgElement = this.mapContainer.getElementsByTagName('svg')[0];
            const labelForEvent = (e) => {
                const x = e.clientX;
                const y = e.clientY;

                let closestDist = Infinity;
                let closest = this.textLabels[0];
                this.textLabels.forEach((label) => {
                    const rect1 = label.getBoundingClientRect();
                    const distance = Math.pow(rect1.left + rect1.width / 2 - x, 2) +
                        Math.pow(rect1.top + rect1.height / 2 - y, 2);
                    if (distance < closestDist) {
                        closestDist = distance;
                        closest = label;
                    }
                });
                return closest;
            };
            let oldClosest = null;
            svgElement.addEventListener('mousemove', (e) => {
                if (oldClosest) {
                    oldClosest.style.fill = 'black';
                }
                oldClosest = labelForEvent(e);
                oldClosest.style.fill = 'blue';
            });
            svgElement.addEventListener('click', (e) => {
                onChosen(labelForEvent(e).textContent);
            });
        }
    }

    function currentStore() {
        const params = new URLSearchParams(location.search);
        return params.get('store');
    }

    function showProductInfo(info) {
        const container = document.createElement('div');
        container.className = 'product-popup';
        container.innerHTML = '<div class="scrollable">' +
            '<label class="name"></label>' +
            '<label class="price"></label>' +
            '<span class="description"></span>' +
            '</div>' +
            '<button class="close-button">Close</button>';

        container.getElementsByClassName('name')[0].textContent = info.name;
        container.getElementsByClassName('price')[0].textContent = info.price;
        container.getElementsByClassName('description')[0].textContent = info.description;

        showPopupDialog(container);
    }

    function setupStoreHeader() {
        const header = document.getElementById('store-header');
        const logo = header.getElementsByClassName('logo')[0];
        const name = header.getElementsByClassName('name')[0];
        logo.src = 'svg/logo/' + window.STORE_DATA.source + '.svg';
        name.textContent = window.STORE_DATA.name;
    }

    window.addEventListener('load', () => {
        window.listPage = new ListPage();
        setupStoreHeader();
    });

})();
