describe("ConnectorDB admin user", function () {
    var cdb;
    beforeEach(function () {
        cdb = new ConnectorDB("test", "test", "http://localhost:8000");
    });

    it("should be able to read user", function (done) {
        cdb.readUser("test").then(function (result) {
            expect(result.admin).toBe(true);
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done)
    });

    it("should be able to create user", function (done) {
        cdb.createUser("javascript_test", "javacript@localhost", "mypass").then(function (result) {
            expect(result.name).toBe("javascript_test");
        }).catch(function (error) {

            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to update user", function (done) {
        cdb.updateUser("javascript_test", {"admin": true}).then(function (result) {
            expect(result.admin).toBe(true);

            return cdb.updateUser("javascript_test", { "admin": false })
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });


    it("should be able to create device", function (done) {
        cdb.createDevice("javascript_test","testdevice").then(function (result) {
            expect(result.name).toBe("testdevice");
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to update device", function (done) {
        cdb.updateDevice("javascript_test", "testdevice", { "nickname": "lolcat" }).then(function (result) {
            expect(result.nickname).toBe("lolcat");
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to read device", function (done) {
        cdb.readDevice("javascript_test", "testdevice").then(function (result) {
            expect(result.nickname).toBe("lolcat");
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to list devices", function (done) {
        cdb.listDevices("javascript_test").then(function (result) {
            expect(result.length).toBe(2);
            expect(result[0].name).toBe("user");
            expect(result[1].name).toBe("testdevice");
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to create stream", function (done) {
        cdb.createStream("javascript_test", "testdevice","mystream",{"type": "boolean"}).then(function (result) {
            expect(result.name).toBe("mystream");
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to update stream", function (done) {
        cdb.updateStream("javascript_test", "testdevice","mystream",{"downlink": true}).then(function (result) {
            expect(result.downlink).toBe(true);

            return cdb.updateStream("javascript_test", "testdevice","mystream",{"downlink": false})
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to read stream", function (done) {
        cdb.readStream("javascript_test", "testdevice", "mystream").then(function (result) {
            expect(result.downlink).toBeUndefined();
            expect(result.name).toBe("mystream")
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to list streams", function (done) {
        cdb.listStreams("javascript_test","testdevice").then(function (result) {
            expect(result.length).toBe(1);
            expect(result[0].name).toBe("mystream");
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to insert into stream", function (done) {
        cdb.insertStream("javascript_test", "testdevice", "mystream",true).then(function (result) {
            expect(result).toBe("ok");
            return cdb.insertStream("javascript_test", "testdevice", "mystream", false)
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to get stream length", function (done) {
        cdb.lengthStream("javascript_test", "testdevice", "mystream").then(function (result) {
            expect(result).toBe(2);
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to read index range", function (done) {
        cdb.indexStream("javascript_test", "testdevice", "mystream",0,5).then(function (result) {
            expect(result.length).toBe(2);
            expect(result[0].d).toBe(true);
            expect(result[1].d).toBe(false);
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to read time range", function (done) {
        cdb.timeStream("javascript_test", "testdevice", "mystream", (new Date).getTime() * 0.001 - 0.5, (new Date).getTime() * 0.001).then(function (result) {
            expect(result.length).toBe(2);
            expect(result[0].d).toBe(true);
            expect(result[1].d).toBe(false);
            return cdb.timeStream("javascript_test", "testdevice", "mystream", (new Date).getTime() * 0.001 - 0.5, (new Date).getTime() * 0.001, 1);
        }).then(function (result) {
            expect(result.length).toBe(1);
            expect(result[0].d).toBe(true);
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to delete stream", function (done) {
        cdb.deleteStream("javascript_test", "testdevice","mystream").then(function (result) {
            expect(result).toBe("ok");
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to delete device", function (done) {
        cdb.deleteDevice("javascript_test", "testdevice").then(function (result) {
            expect(result).toBe("ok");
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });

    it("should be able to delete user", function (done) {
        cdb.deleteUser("javascript_test").then(function (result) {
            expect(result).toBe("ok");
        }).catch(function (error) {
            expect(error).toBeUndefined();
        }).then(done);
    });
});
