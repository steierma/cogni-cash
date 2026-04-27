export const getClientId = (): string => {
    let id = localStorage.getItem('cc_client_id');
    if (!id) {
        id = crypto.randomUUID();
        localStorage.setItem('cc_client_id', id);
    }
    return id;
};
