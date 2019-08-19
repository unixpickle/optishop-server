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
        }

        hideList() {
            super.hideList();
            this.totalPrice.style.display = 'none';
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
            const formData = 'store=' + encodeURIComponent(currentStore()) +
                '&signature=' + encodeURIComponent(item.signature) +
                '&data=' + encodeURIComponent(JSON.stringify(item.data));
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
            if (data.error) {
                throw data.error;
            }
            this.updateData(data);
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
            return result.filter((x) => x.inStock);
        }

        createListItem(item) {
            return createListItem(item);
        }
    }

    function currentStore() {
        const params = new URLSearchParams(location.search);
        return params.get('store');
    }

    function createListItem(item) {
        const elem = document.createElement('li');
        elem.className = 'list-item';

        const image = document.createElement('img');
        image.className = 'image';
        image.src = item.photoUrl;
        elem.appendChild(image);

        const name = document.createElement('label');
        name.className = 'name';
        name.textContent = item.name;
        elem.appendChild(name);

        const zone = document.createElement('label');
        zone.className = 'location';
        if (item.zone) {
            zone.textContent = item.zone;
        } else {
            // For the search screen.
            zone.textContent = item.price;
        }
        elem.appendChild(zone);

        return elem;
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
