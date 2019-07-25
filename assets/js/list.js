(function () {

    class ListPage extends ListingPage {
        constructor() {
            super();
            this.sortButton = document.getElementById('sort-button');
            this.sortButton.addEventListener('click', () => {
                const hideLoader = showOverlayLoader();
                this.sort().catch(handleError).finally(hideLoader);
            });
            this.routeButton = document.getElementById('route-button');
            this.routeButton.addEventListener('click', () => {
                window.open('/route?store=' + encodeURIComponent(currentStore()));
            });
            this.updateData(this.data || window.LIST_DATA);
        }

        async sort() {
            const response = await fetch('/api/sort?store=' + encodeURIComponent(currentStore()), {
                credentials: 'same-origin',
            });
            const data = await response.json();
            if (data.error) {
                throw data.error;
            }
            this.updateData(data);
        }

        createAddDialog() {
            return new AddProductDialog();
        }

        createListItem(store) {
            return createListItem(store);
        }

        selectedListItem(store) {
            // TODO: show product info dialog here.
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
