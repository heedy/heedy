/**
 * This file is the test suite for the type-checking code used to ensure that 
 * a given datapoint array can be displayed by a certain visualization
 */

import * as tc from '../js/datatypes/views/typecheck';

let dpa = [{ t: 11, d: 23 }];

describe("Datapoint Array Type Checking", function () {
    describe("single datapoint type check", function () {
        it("checks isNumber", function () {
            expect(tc.isNumber(0.3)).toBe(true);
            expect(tc.isNumber(-3453)).toBe(true);
            expect(tc.isNumber("1337.6")).toBe(true);
            expect(tc.isNumber(true)).toBe(false);
            expect(tc.isNumber("hi")).toBe(false);
        });
        it("checks isString", function () {
            expect(tc.isString(0.3)).toBe(false);
            expect(tc.isString("-3453")).toBe(true);
            expect(tc.isString("")).toBe(true);
        });
        it("checks isBool", function () {
            expect(tc.isBool(0.3)).toBe(false);
            expect(tc.isBool("true")).toBe(true);
            expect(tc.isBool("")).toBe(false);
        });
    });

});