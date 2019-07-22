const params = new URLSearchParams(location.search);
if (params.get('error')) {
    const errBox = document.getElementById('account-error');
    errBox.textContent = params.get('error');
    errBox.style.display = 'block';
}
