(function () {

    class StoresPage extends ListingPage {
        async fetchData() {
            const response = await fetch('/api/stores', {
                credentials: 'same-origin',
                cache: 'no-store',
            });
            const data = await response.json();
            if (data.error) {
                throw data.error;
            }
            return data;
        }

        createAddDialog() {
            return new AddStoreDialog();
        }

        createListItem(store) {
            return createListItem(store);
        }

        selectedListItem(store) {
            window.location = '/list?store=' + encodeURIComponent(store.id);
        }

        async addItem(store) {
            const formData = 'source=' + encodeURIComponent(store.source) +
                '&signature=' + encodeURIComponent(store.signature) +
                '&data=' + encodeURIComponent(JSON.stringify(store.data));
            const response = await fetch('/api/addstore', {
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
            store.id = data;
            this.addListItem(store);
            this.addDialog.close();
        }

        async deleteItem(store) {
            const response = await fetch('/api/removestore?store=' + encodeURIComponent(store.id), {
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

    class AddStoreDialog extends AddDialog {
        async fetchSearchResults(query) {
            const response = await fetch('/api/storequery?query=' + encodeURIComponent(query), {
                credentials: 'same-origin',
                cache: 'no-store',
            });
            const result = await response.json();
            if (result.error) {
                throw result.error;
            }
            return result;
        }

        createListItem(item) {
            return createListItem(item);
        }
    }

    function createListItem(store) {
        const elem = document.createElement('li');
        elem.className = 'list-item';

        const logo = document.createElement('img');
        logo.className = 'image';
        logo.src = 'svg/logo/' + store.source + '.svg';
        elem.appendChild(logo);

        const name = document.createElement('label');
        name.className = 'name';
        name.textContent = store.name;
        elem.appendChild(name);

        const address = document.createElement('label');
        address.className = 'location';
        address.textContent = store.address;
        elem.appendChild(address);

        return elem;
    }

    window.addEventListener('load', () => {
        window.storesPage = new StoresPage();
    });

})();
