/*
Sagas are used to asynchronously execute actions upon events.

TODO: Up until now, most of ConnectorDB was built with actions hacked on wherever possible.
This led to a tough codebase, and a rather unpleasant difficulty of doing anything interesting.
With Saga, this difficulty is gone. Unfortunately, most of the core frontend code is still in the hacky format
used before saga. At some point, this code should be cleaned up and converted into sagas.

Since fixing old-but-working code is not as big a priority as getting the functionality working,
all more recent code uses newer coding practices and the knowledge gained from the downfalls of the old code.

That's why the codebase has inconsistent practices - with some code being almost entirely functional with sagas,
and other code being... not.
*/
import downlinkSaga from './downlinks';
import analysisSaga from './analysis';
import navigationSaga from './navigation';

export default function* sagas() {
    yield [
        downlinkSaga(),
        analysisSaga(),
        navigationSaga()
    ];
}