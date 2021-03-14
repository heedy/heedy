from .base import Session, q


class KV:
    def __init__(self, uri: str, session: Session):
        self.session = session
        self.uri = uri

    def _uri(self, namespace):
        if namespace is None:
            namespace = self.session.namespace
        return self.uri + "/" + q(namespace)

    def get(self, namespace=None):
        return self.session.get(self._uri(namespace))

    def getkey(self, key: str, namespace=None):
        return self.session.get(self._uri(namespace) + "/" + q(key))

    def set(self, namespace=None, **kwargs):
        return self.session.post(self._uri(namespace), kwargs)

    def update(self, namespace=None, **kwargs):
        return self.session.patch(self._uri(namespace), kwargs)

    def delete(self, key, namespace=None):
        return self.session.delete(self._uri(namespace) + "/" + q(key))

    def __getitem__(self, key: str):
        return self.getkey(key)

    def __setitem__(self, key: str, value):
        return self.set(**{key: value})

    def __delitem__(self, key: str):
        return self.delete(key)

    def __call__(self, namespace=None):
        return self.get(namespace)
