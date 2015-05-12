var expect = chai.expect;

describe("ConnectorDB", function() {
    it("should get this correctly", function () {
        var db = new ConnectorDB("test", "test", "localhost:8000")
        expect(db.thisdevice()).to.equal("https://connectordb.com")
    })
})