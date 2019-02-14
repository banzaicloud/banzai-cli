(window["webpackJsonp"] = window["webpackJsonp"] || []).push([["main"],{

/***/ "./dist/banzaicloud/uniform/fesm5/banzaicloud-uniform.js":
/*!***************************************************************!*\
  !*** ./dist/banzaicloud/uniform/fesm5/banzaicloud-uniform.js ***!
  \***************************************************************/
/*! exports provided: FormFieldCheckbox, FormFieldCode, FormFieldFile, FormFieldText, FormFieldNumber, FormFieldPassword, FormFieldSelect, FormFieldTextarea, CONTROL_TYPES, CONTROL_GROUP_TYPES, UniformModule, UniformService, ɵi, ɵk, ɵp, ɵf, ɵa, ɵl, ɵg, ɵo, ɵm, ɵn, ɵc, ɵe, ɵd, ɵb, ɵj, ɵh */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "FormFieldCheckbox", function() { return FormFieldCheckbox; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "FormFieldCode", function() { return FormFieldCode; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "FormFieldFile", function() { return FormFieldFile; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "FormFieldText", function() { return FormFieldText; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "FormFieldNumber", function() { return FormFieldNumber; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "FormFieldPassword", function() { return FormFieldPassword; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "FormFieldSelect", function() { return FormFieldSelect; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "FormFieldTextarea", function() { return FormFieldTextarea; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "CONTROL_TYPES", function() { return CONTROL_TYPES; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "CONTROL_GROUP_TYPES", function() { return CONTROL_GROUP_TYPES; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "UniformModule", function() { return UniformModule; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "UniformService", function() { return UniformService; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵi", function() { return DisableControlDirective; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵk", function() { return FormFieldContainerComponent; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵp", function() { return FormField; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵf", function() { return FormFieldCheckboxComponent; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵa", function() { return FormFieldCodeComponent; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵl", function() { return FileValueAccessorDirective; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵg", function() { return FormFieldFileComponent; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵo", function() { return MatFileInputDirective; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵm", function() { return MatInputBase; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵn", function() { return _MatInputMixinBase; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵc", function() { return FormFieldInputComponent; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵe", function() { return FormFieldSelectComponent; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵd", function() { return FormFieldTextareaComponent; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵb", function() { return FormFieldComponent; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵj", function() { return FormGroupComponent; });
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "ɵh", function() { return FormComponent; });
/* harmony import */ var _angular_common__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! @angular/common */ "./node_modules/@angular/common/fesm5/common.js");
/* harmony import */ var _angular_platform_browser_animations__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/platform-browser/animations */ "./node_modules/@angular/platform-browser/fesm5/animations.js");
/* harmony import */ var _angular_material_input__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! @angular/material/input */ "./node_modules/@angular/material/esm5/input.es5.js");
/* harmony import */ var _angular_material_checkbox__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! @angular/material/checkbox */ "./node_modules/@angular/material/esm5/checkbox.es5.js");
/* harmony import */ var _angular_material_select__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! @angular/material/select */ "./node_modules/@angular/material/esm5/select.es5.js");
/* harmony import */ var _angular_material_button__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! @angular/material/button */ "./node_modules/@angular/material/esm5/button.es5.js");
/* harmony import */ var _angular_flex_layout__WEBPACK_IMPORTED_MODULE_6__ = __webpack_require__(/*! @angular/flex-layout */ "./node_modules/@angular/flex-layout/esm5/flex-layout.es5.js");
/* harmony import */ var _ctrl_ngx_codemirror__WEBPACK_IMPORTED_MODULE_7__ = __webpack_require__(/*! @ctrl/ngx-codemirror */ "./node_modules/@ctrl/ngx-codemirror/fesm5/ctrl-ngx-codemirror.js");
/* harmony import */ var codemirror_mode_javascript_javascript__WEBPACK_IMPORTED_MODULE_8__ = __webpack_require__(/*! codemirror/mode/javascript/javascript */ "./node_modules/codemirror/mode/javascript/javascript.js");
/* harmony import */ var codemirror_mode_javascript_javascript__WEBPACK_IMPORTED_MODULE_8___default = /*#__PURE__*/__webpack_require__.n(codemirror_mode_javascript_javascript__WEBPACK_IMPORTED_MODULE_8__);
/* harmony import */ var codemirror_mode_markdown_markdown__WEBPACK_IMPORTED_MODULE_9__ = __webpack_require__(/*! codemirror/mode/markdown/markdown */ "./node_modules/codemirror/mode/markdown/markdown.js");
/* harmony import */ var codemirror_mode_markdown_markdown__WEBPACK_IMPORTED_MODULE_9___default = /*#__PURE__*/__webpack_require__.n(codemirror_mode_markdown_markdown__WEBPACK_IMPORTED_MODULE_9__);
/* harmony import */ var codemirror_mode_yaml_yaml__WEBPACK_IMPORTED_MODULE_10__ = __webpack_require__(/*! codemirror/mode/yaml/yaml */ "./node_modules/codemirror/mode/yaml/yaml.js");
/* harmony import */ var codemirror_mode_yaml_yaml__WEBPACK_IMPORTED_MODULE_10___default = /*#__PURE__*/__webpack_require__.n(codemirror_mode_yaml_yaml__WEBPACK_IMPORTED_MODULE_10__);
/* harmony import */ var codemirror_addon_display_placeholder__WEBPACK_IMPORTED_MODULE_11__ = __webpack_require__(/*! codemirror/addon/display/placeholder */ "./node_modules/codemirror/addon/display/placeholder.js");
/* harmony import */ var codemirror_addon_display_placeholder__WEBPACK_IMPORTED_MODULE_11___default = /*#__PURE__*/__webpack_require__.n(codemirror_addon_display_placeholder__WEBPACK_IMPORTED_MODULE_11__);
/* harmony import */ var rxjs_operators__WEBPACK_IMPORTED_MODULE_12__ = __webpack_require__(/*! rxjs/operators */ "./node_modules/rxjs/_esm5/operators/index.js");
/* harmony import */ var _angular_material_core__WEBPACK_IMPORTED_MODULE_13__ = __webpack_require__(/*! @angular/material/core */ "./node_modules/@angular/material/esm5/core.es5.js");
/* harmony import */ var _angular_material__WEBPACK_IMPORTED_MODULE_14__ = __webpack_require__(/*! @angular/material */ "./node_modules/@angular/material/esm5/material.es5.js");
/* harmony import */ var rxjs__WEBPACK_IMPORTED_MODULE_15__ = __webpack_require__(/*! rxjs */ "./node_modules/rxjs/_esm5/index.js");
/* harmony import */ var _angular_cdk_coercion__WEBPACK_IMPORTED_MODULE_16__ = __webpack_require__(/*! @angular/cdk/coercion */ "./node_modules/@angular/cdk/esm5/coercion.es5.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_17__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var lodash_es__WEBPACK_IMPORTED_MODULE_18__ = __webpack_require__(/*! lodash-es */ "./node_modules/lodash-es/lodash.js");
/* harmony import */ var ajv__WEBPACK_IMPORTED_MODULE_19__ = __webpack_require__(/*! ajv */ "./node_modules/ajv/lib/ajv.js");
/* harmony import */ var ajv__WEBPACK_IMPORTED_MODULE_19___default = /*#__PURE__*/__webpack_require__.n(ajv__WEBPACK_IMPORTED_MODULE_19__);
/* harmony import */ var _angular_forms__WEBPACK_IMPORTED_MODULE_20__ = __webpack_require__(/*! @angular/forms */ "./node_modules/@angular/forms/fesm5/forms.js");
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_21__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");























/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
/** @enum {string} */
var CONTROL_TYPES = {
    CHECKBOX: 'checkbox',
    CODE: 'code',
    FILE: 'file',
    NUMBER: 'number',
    PASSWORD: 'password',
    SELECT: 'select',
    TEXT: 'text',
    TEXTAREA: 'textarea',
};
/** @enum {string} */
var CONTROL_GROUP_TYPES = {
    AMAZON: 'amazon',
    AZURE: 'azure',
    ALIBABA: 'alibaba',
    GOOGLE: 'google',
    ORACLE: 'oracle',
    PASSWORD: 'password',
    TLS: 'tls',
};

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldComponent = /** @class */ (function () {
    function FormFieldComponent() {
    }
    Object.defineProperty(FormFieldComponent.prototype, "control", {
        get: /**
         * @return {?}
         */
        function () {
            return this.form.controls[this.field.key];
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(FormFieldComponent.prototype, "isValid", {
        get: /**
         * @return {?}
         */
        function () {
            return this.control.valid;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(FormFieldComponent.prototype, "isTouched", {
        get: /**
         * @return {?}
         */
        function () {
            return this.control.touched;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(FormFieldComponent.prototype, "errorMessage", {
        get: /**
         * @return {?}
         */
        function () {
            var errors = this.control.errors;
            if (errors) {
                if (errors.required) {
                    return 'The field is required';
                }
                if (errors.minlength) {
                    return "The field must be at least " + errors.minlength.requiredLength + " characters long";
                }
                if (errors.maxlength) {
                    return "The field must be at most " + errors.maxlength.requiredLength + " characters long";
                }
                if (errors.min) {
                    return "The field must be more or equal to " + errors.min.min;
                }
                if (errors.max) {
                    return "The field must be less or equal to " + errors.max.max;
                }
                if (errors.pattern) {
                    return "The field must match the '" + errors.pattern.requiredPattern + "' pattern";
                }
                if (errors.custom) {
                    return errors.custom;
                }
                return "Error: " + JSON.stringify(errors);
            }
            return '';
        },
        enumerable: true,
        configurable: true
    });
    FormFieldComponent.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Component"], args: [{
                    selector: 'bcu-form-field',
                    template: '',
                    encapsulation: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ViewEncapsulation"].None
                }] }
    ];
    FormFieldComponent.propDecorators = {
        field: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        form: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }]
    };
    return FormFieldComponent;
}());

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormComponent = /** @class */ (function () {
    function FormComponent() {
        this.fields = [];
        this.form = new _angular_forms__WEBPACK_IMPORTED_MODULE_20__["FormGroup"]({}, {});
        this.groups = [];
        this.save = new _angular_core__WEBPACK_IMPORTED_MODULE_17__["EventEmitter"]();
        this.valueChanges = new _angular_core__WEBPACK_IMPORTED_MODULE_17__["EventEmitter"]();
        this.destroy$ = new rxjs__WEBPACK_IMPORTED_MODULE_15__["Subject"]();
    }
    Object.defineProperty(FormComponent.prototype, "fieldsByKey", {
        get: /**
         * @private
         * @return {?}
         */
        function () {
            return this.fields.reduce((/**
             * @param {?} all
             * @param {?} field
             * @return {?}
             */
            function (all, field) {
                var _a;
                return (Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__assign"])({}, all, (_a = {}, _a[field.key] = field, _a)));
            }), {});
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(FormComponent.prototype, "_groups", {
        set: /**
         * @param {?} groups
         * @return {?}
         */
        function (groups) {
            if (!groups) {
                return;
            }
            this.groups = groups;
            this.fields = Object(lodash_es__WEBPACK_IMPORTED_MODULE_18__["flatten"])(this.groups.map((/**
             * @param {?} group
             * @return {?}
             */
            function (group) { return group.fields; })));
        },
        enumerable: true,
        configurable: true
    });
    /**
     * @return {?}
     */
    FormComponent.prototype.ngOnInit = /**
     * @return {?}
     */
    function () {
        /** @type {?} */
        var controls = Object(lodash_es__WEBPACK_IMPORTED_MODULE_18__["mapValues"])(this.fieldsByKey, (/**
         * @param {?} __0
         * @return {?}
         */
        function (_a) {
            var control = _a.control;
            return control;
        }));
        // set initial state
        if (this.initialValue) {
            Object(lodash_es__WEBPACK_IMPORTED_MODULE_18__["forEach"])(this.initialValue, (/**
             * @param {?} val
             * @param {?} key
             * @return {?}
             */
            function (val, key) {
                /** @type {?} */
                var control = controls[key];
                if (control) {
                    control.setValue(val);
                }
            }));
        }
        this.form = new _angular_forms__WEBPACK_IMPORTED_MODULE_20__["FormGroup"](controls, {});
        this._valueChangesSubscription = this.form.valueChanges.pipe(Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_12__["takeUntil"])(this.destroy$)).subscribe(this.valueChanges);
    };
    /**
     * @param {?} changes
     * @return {?}
     */
    FormComponent.prototype.ngOnChanges = /**
     * @param {?} changes
     * @return {?}
     */
    function (changes) {
        var _this = this;
        if (changes['_groups'] && !changes['_groups'].isFirstChange()) {
            /** @type {?} */
            var controls_1 = Object(lodash_es__WEBPACK_IMPORTED_MODULE_18__["mapValues"])(this.fieldsByKey, (/**
             * @param {?} __0
             * @return {?}
             */
            function (_a) {
                var control = _a.control;
                return control;
            }));
            // keep current save
            Object(lodash_es__WEBPACK_IMPORTED_MODULE_18__["forEach"])(this.form.value, (/**
             * @param {?} val
             * @param {?} key
             * @return {?}
             */
            function (val, key) {
                /** @type {?} */
                var newControl = controls_1[key];
                /** @type {?} */
                var currentControl = _this.form.controls[key];
                if (newControl) {
                    newControl.setValue(val);
                    if (currentControl.touched) {
                        newControl.markAsTouched({ onlySelf: true });
                    }
                }
            }));
            if (this._valueChangesSubscription) {
                this._valueChangesSubscription.unsubscribe();
            }
            this.form = new _angular_forms__WEBPACK_IMPORTED_MODULE_20__["FormGroup"](controls_1, {});
            this._valueChangesSubscription = this.form.valueChanges
                .pipe(Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_12__["takeUntil"])(this.destroy$))
                .subscribe(this.valueChanges);
            this.valueChanges.emit(this.form.value);
        }
    };
    /**
     * @return {?}
     */
    FormComponent.prototype.ngOnDestroy = /**
     * @return {?}
     */
    function () {
        this.destroy$.next();
        this.destroy$.complete();
    };
    /**
     * @return {?}
     */
    FormComponent.prototype.onReset = /**
     * @return {?}
     */
    function () {
        /** @type {?} */
        var defaults = Object(lodash_es__WEBPACK_IMPORTED_MODULE_18__["mapValues"])(this.fieldsByKey, (/**
         * @param {?} __0
         * @return {?}
         */
        function (_a) {
            var defaultValue = _a.default;
            return defaultValue;
        }));
        this.form.reset(defaults, { onlySelf: true });
    };
    /**
     * @return {?}
     */
    FormComponent.prototype.onSave = /**
     * @return {?}
     */
    function () {
        if (this.form.valid) {
            this.save.emit(this.form.getRawValue());
        }
        else {
            Object(lodash_es__WEBPACK_IMPORTED_MODULE_18__["forEach"])(this.form.controls, (/**
             * @param {?} control
             * @return {?}
             */
            function (control) { return control.markAsTouched({ onlySelf: true }); }));
        }
    };
    FormComponent.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Component"], args: [{
                    selector: 'bcu-form',
                    template: "<form [formGroup]=\"form\" fxLayout=\"column\">\n  <bcu-form-group\n    *ngFor=\"let group of groups\"\n    [form]=\"form\"\n    [name]=\"group.name\"\n    [description]=\"group.description\"\n    [link]=\"group.link\"\n    [fields]=\"group.fields\"\n  ></bcu-form-group>\n  <div fxLayout=\"row\">\n    <button fxFlex mat-button (click)=\"onReset()\">Clear</button>\n    <button fxFlex mat-button (click)=\"onSave()\">Save</button>\n  </div>\n</form>\n",
                    encapsulation: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ViewEncapsulation"].None,
                    styles: ["bcu-form-group:not(:first-child){margin-top:32px}"]
                }] }
    ];
    FormComponent.propDecorators = {
        _groups: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"], args: ['groups',] }],
        initialValue: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        save: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Output"] }],
        valueChanges: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Output"] }]
    };
    return FormComponent;
}());

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var DisableControlDirective = /** @class */ (function () {
    function DisableControlDirective(ngControl) {
        this.ngControl = ngControl;
    }
    Object.defineProperty(DisableControlDirective.prototype, "bcuDisableControl", {
        set: /**
         * @param {?} condition
         * @return {?}
         */
        function (condition) {
            var _this = this;
            setTimeout((/**
             * @return {?}
             */
            function () {
                if (condition) {
                    _this.ngControl.control.disable();
                }
                else {
                    _this.ngControl.control.enable();
                }
            }));
        },
        enumerable: true,
        configurable: true
    });
    DisableControlDirective.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Directive"], args: [{
                    selector: '[bcuDisableControl]',
                },] }
    ];
    /** @nocollapse */
    DisableControlDirective.ctorParameters = function () { return [
        { type: _angular_forms__WEBPACK_IMPORTED_MODULE_20__["NgControl"] }
    ]; };
    DisableControlDirective.propDecorators = {
        bcuDisableControl: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }]
    };
    return DisableControlDirective;
}());

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormGroupComponent = /** @class */ (function () {
    function FormGroupComponent() {
    }
    FormGroupComponent.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Component"], args: [{
                    selector: 'bcu-form-group',
                    template: "<div class=\"bcu-form-group\">\n  <h3>{{ name }}</h3>\n  <p *ngIf=\"description\">{{ description }}</p>\n  <a *ngIf=\"link\" [href]=\"link\" target=\"_blank\" rel=\"noopener noreferrer\">{{ link }}</a>\n  <div class=\"bcu-form-group-fields\">\n    <ng-container *ngFor=\"let field of fields\">\n      <bcu-form-field-container\n        [ngClass]=\"{ hidden: field.isHidden(form.value) }\"\n        [form]=\"form\"\n        [field]=\"field\"\n      ></bcu-form-field-container>\n    </ng-container>\n  </div>\n</div>\n",
                    encapsulation: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ViewEncapsulation"].None,
                    styles: [".bcu-form-group h3{font-size:16px;font-weight:700;line-height:1.5;letter-spacing:.2px;height:24px;margin:0 0 16px}.bcu-form-group p{padding:0;margin:0}.bcu-form-group a{font-size:75%;line-height:1.5;letter-spacing:.2px;margin:0;padding:0}.bcu-form-group .bcu-form-group-fields{margin-top:16px}.bcu-form-group .hidden{display:none}"]
                }] }
    ];
    FormGroupComponent.propDecorators = {
        form: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        name: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        description: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        link: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        fields: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }]
    };
    return FormGroupComponent;
}());

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldContainerComponent = /** @class */ (function () {
    function FormFieldContainerComponent(componentFactoryResolver) {
        this.componentFactoryResolver = componentFactoryResolver;
    }
    /**
     * @return {?}
     */
    FormFieldContainerComponent.prototype.ngOnInit = /**
     * @return {?}
     */
    function () {
        if (this.field.component) {
            /** @type {?} */
            var componentFactory = this.componentFactoryResolver.resolveComponentFactory(this.field.component);
            this.containerRef.clear();
            this.componentRef = this.containerRef.createComponent(componentFactory);
            if (this.componentRef.instance) {
                Object.assign(this.componentRef.instance, {
                    form: this.form,
                    field: this.field,
                });
            }
        }
    };
    /**
     * @return {?}
     */
    FormFieldContainerComponent.prototype.ngOnChanges = /**
     * @return {?}
     */
    function () {
        if (this.componentRef && this.componentRef.instance) {
            Object.assign(this.componentRef.instance, {
                form: this.form,
                field: this.field,
            });
        }
    };
    FormFieldContainerComponent.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Component"], args: [{
                    selector: 'bcu-form-field-container',
                    template: "<div class=\"bcu-form-field-container\">\n  <section fxLayout=\"column\">\n    <ng-container #container></ng-container>\n  </section>\n</div>\n",
                    encapsulation: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ViewEncapsulation"].None,
                    styles: [".bcu-form-field-container label{text-transform:uppercase}.bcu-form-field-container .mat-form-field{margin-top:16px;line-height:1.5;letter-spacing:.2px}"]
                }] }
    ];
    /** @nocollapse */
    FormFieldContainerComponent.ctorParameters = function () { return [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ComponentFactoryResolver"] }
    ]; };
    FormFieldContainerComponent.propDecorators = {
        field: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        form: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        containerRef: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ViewChild"], args: ['container', { read: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ViewContainerRef"] },] }]
    };
    return FormFieldContainerComponent;
}());

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldCodeComponent = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldCodeComponent, _super);
    function FormFieldCodeComponent() {
        return _super !== null && _super.apply(this, arguments) || this;
    }
    Object.defineProperty(FormFieldCodeComponent.prototype, "options", {
        get: /**
         * @return {?}
         */
        function () {
            return Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__assign"])({ lineNumbers: true, theme: 'material', mode: { name: 'javascript', json: true }, viewportMargin: 100, matchBrackets: true, lineWrapping: true }, this.field.options, { readOnly: this.field.isDisabled(this.form.value), placeholder: this.field.placeholder });
        },
        enumerable: true,
        configurable: true
    });
    FormFieldCodeComponent.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Component"], args: [{
                    selector: 'bcu-form-field-code',
                    template: "<div [formGroup]=\"form\" class=\"bcu-form-field-code\" fxFlex>\n  <mat-form-field floatLabel=\"always\" class=\"mat-form-field--no-underline\" fxFlex>\n    <input matInput [id]=\"field.key\" [formControlName]=\"field.key\" [required]=\"field.required\" style=\"display: none\" />\n    <mat-label>{{ field.label }}</mat-label>\n\n    <ngx-codemirror [formControlName]=\"field.key\" [options]=\"options\"></ngx-codemirror>\n\n    <mat-hint *ngIf=\"field.description\" align=\"start\">{{ field.description }}</mat-hint>\n    <mat-error *ngIf=\"isTouched && !isValid\">{{ errorMessage }}</mat-error>\n  </mat-form-field>\n</div>\n",
                    encapsulation: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ViewEncapsulation"].None,
                    styles: [".bcu-form-field-code .mat-form-field--no-underline .mat-form-field-ripple,.bcu-form-field-code .mat-form-field--no-underline .mat-form-field-underline{background-color:transparent}"]
                }] }
    ];
    FormFieldCodeComponent.propDecorators = {
        field: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }]
    };
    return FormFieldCodeComponent;
}(FormFieldComponent));

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldInputComponent = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldInputComponent, _super);
    function FormFieldInputComponent() {
        return _super.call(this) || this;
    }
    FormFieldInputComponent.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Component"], args: [{
                    selector: 'bcu-form-field-input',
                    template: "<div [formGroup]=\"form\" fxFlex>\n  <mat-form-field floatLabel=\"always\" fxFlex=\"\">\n    <mat-label>{{ field.label }}</mat-label>\n    <input\n      matInput\n      [type]=\"field.controlType\"\n      [formControlName]=\"field.key\"\n      [id]=\"field.key\"\n      [required]=\"field.required\"\n      [placeholder]=\"field.placeholder\"\n      [bcuDisableControl]=\"field.isDisabled(form.value)\"\n    />\n\n    <mat-hint *ngIf=\"field.description\" align=\"start\">{{ field.description }}</mat-hint>\n    <mat-error *ngIf=\"isTouched && !isValid\">{{ errorMessage }}</mat-error>\n  </mat-form-field>\n</div>\n",
                    encapsulation: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ViewEncapsulation"].None,
                    styles: [""]
                }] }
    ];
    /** @nocollapse */
    FormFieldInputComponent.ctorParameters = function () { return []; };
    return FormFieldInputComponent;
}(FormFieldComponent));

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldTextareaComponent = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldTextareaComponent, _super);
    function FormFieldTextareaComponent() {
        return _super.call(this) || this;
    }
    FormFieldTextareaComponent.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Component"], args: [{
                    selector: 'bcu-form-field-textarea',
                    template: "<div [formGroup]=\"form\" fxFlex>\n  <mat-form-field floatLabel=\"always\" fxFlex>\n    <mat-label>{{ field.label }}</mat-label>\n\n    <textarea\n      matInput\n      [formControlName]=\"field.key\"\n      [id]=\"field.key\"\n      [required]=\"field.required\"\n      [placeholder]=\"field.placeholder\"\n      [bcuDisableControl]=\"field.isDisabled(form.value)\"\n    ></textarea>\n\n    <mat-hint *ngIf=\"field.description\" align=\"start\">{{ field.description }}</mat-hint>\n    <mat-error *ngIf=\"isTouched && !isValid\">{{ errorMessage }}</mat-error>\n  </mat-form-field>\n</div>\n",
                    encapsulation: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ViewEncapsulation"].None,
                    styles: [""]
                }] }
    ];
    /** @nocollapse */
    FormFieldTextareaComponent.ctorParameters = function () { return []; };
    return FormFieldTextareaComponent;
}(FormFieldComponent));

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldSelectComponent = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldSelectComponent, _super);
    function FormFieldSelectComponent() {
        return _super.call(this) || this;
    }
    FormFieldSelectComponent.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Component"], args: [{
                    selector: 'bcu-form-field-select',
                    template: "<div [formGroup]=\"form\" fxFlex>\n  <mat-form-field floatLabel=\"always\" fxFlex>\n    <mat-label>{{ field.label }}</mat-label>\n    <mat-select\n      [id]=\"field.key\"\n      [formControlName]=\"field.key\"\n      [required]=\"field.required\"\n      [placeholder]=\"field.placeholder\"\n      [bcuDisableControl]=\"field.isDisabled(form.value)\"\n    >\n      <mat-option *ngFor=\"let opt of field.options\" [value]=\"opt.value || opt\">{{ opt.label || opt }}</mat-option>\n    </mat-select>\n\n    <mat-hint *ngIf=\"field.description\" align=\"start\">{{ field.description }}</mat-hint>\n    <mat-error *ngIf=\"isTouched && !isValid\">{{ errorMessage }}</mat-error>\n  </mat-form-field>\n</div>\n",
                    encapsulation: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ViewEncapsulation"].None,
                    styles: [""]
                }] }
    ];
    /** @nocollapse */
    FormFieldSelectComponent.ctorParameters = function () { return []; };
    FormFieldSelectComponent.propDecorators = {
        field: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }]
    };
    return FormFieldSelectComponent;
}(FormFieldComponent));

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldCheckboxComponent = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldCheckboxComponent, _super);
    function FormFieldCheckboxComponent() {
        return _super.call(this) || this;
    }
    FormFieldCheckboxComponent.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Component"], args: [{
                    selector: 'bcu-form-field-checkbox',
                    template: "<div [formGroup]=\"form\" class=\"bcu-form-field-checkbox\" fxFlex>\n  <mat-form-field floatLabel=\"always\" class=\"mat-form-field--no-underline\" fxFlex>\n    <input matInput [id]=\"field.key\" [formControlName]=\"field.key\" [required]=\"field.required\" style=\"display: none\" />\n    <mat-label *ngIf=\"field.placeholder\">{{ field.placeholder }}</mat-label>\n    <mat-checkbox\n      [formControlName]=\"field.key\"\n      [id]=\"field.key\"\n      [required]=\"field.required\"\n      [bcuDisableControl]=\"field.isDisabled(form.value)\"\n    >\n      {{ field.label }}\n    </mat-checkbox>\n\n    <mat-hint *ngIf=\"field.description\" align=\"start\">{{ field.description }}</mat-hint>\n    <mat-error *ngIf=\"isTouched && !isValid\">{{ errorMessage }}</mat-error>\n  </mat-form-field>\n</div>\n",
                    encapsulation: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ViewEncapsulation"].None,
                    styles: [".bcu-form-field-checkbox .mat-form-field--no-underline .mat-form-field-ripple,.bcu-form-field-checkbox .mat-form-field--no-underline .mat-form-field-underline{background-color:transparent}"]
                }] }
    ];
    /** @nocollapse */
    FormFieldCheckboxComponent.ctorParameters = function () { return []; };
    return FormFieldCheckboxComponent;
}(FormFieldComponent));

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldFileComponent = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldFileComponent, _super);
    function FormFieldFileComponent() {
        var _this = _super !== null && _super.apply(this, arguments) || this;
        _this.dragActive = false;
        return _this;
    }
    /**
     * @return {?}
     */
    FormFieldFileComponent.prototype.ngOnChanges = /**
     * @return {?}
     */
    function () {
        if (this.form && this.field && this.field.default) {
            /** @type {?} */
            var control = this.form.controls[this.field.key];
            if (control) {
                control.setValue(this.field.default);
            }
        }
    };
    /**
     * @private
     * @param {?} files
     * @return {?}
     */
    FormFieldFileComponent.prototype.onFile = /**
     * @private
     * @param {?} files
     * @return {?}
     */
    function (files) {
        var _this = this;
        /** @type {?} */
        var file = files[0];
        if (file) {
            /** @type {?} */
            var reader_1 = new FileReader();
            reader_1.onload = (/**
             * @return {?}
             */
            function () {
                var e_1, _a;
                _this.form.controls[_this.field.key].setValue(reader_1.result);
                if (_this.field.fillForm) {
                    /** @type {?} */
                    var baseKey = _this.field.key.split('.')[0];
                    try {
                        /** @type {?} */
                        var result = JSON.parse((/** @type {?} */ (reader_1.result)));
                        try {
                            for (var _b = Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__values"])(Object.keys(result)), _c = _b.next(); !_c.done; _c = _b.next()) {
                                var key = _c.value;
                                /** @type {?} */
                                var control = _this.form.controls[baseKey + "." + key];
                                if (control) {
                                    control.setValue(result[key]);
                                    control.markAsTouched({ onlySelf: true });
                                }
                            }
                        }
                        catch (e_1_1) { e_1 = { error: e_1_1 }; }
                        finally {
                            try {
                                if (_c && !_c.done && (_a = _b.return)) _a.call(_b);
                            }
                            finally { if (e_1) throw e_1.error; }
                        }
                    }
                    catch (err) {
                        /** @type {?} */
                        var control = _this.form.controls[_this.field.key];
                        control.setErrors({ custom: 'Failed to parse file as JSON' });
                        control.markAsTouched({ onlySelf: true });
                        console.log(err);
                    }
                }
            });
            reader_1.readAsText(file);
        }
    };
    /**
     * @param {?} event
     * @return {?}
     */
    FormFieldFileComponent.prototype.onFileSelect = /**
     * @param {?} event
     * @return {?}
     */
    function (event) {
        this.onFile(event.target.files);
    };
    /**
     * @param {?} event
     * @return {?}
     */
    FormFieldFileComponent.prototype.onDrop = /**
     * @param {?} event
     * @return {?}
     */
    function (event) {
        event.preventDefault();
        if (this.field.fillForm && !event.dataTransfer.files.length) {
            /** @type {?} */
            var control = this.form.controls[this.field.key];
            control.setErrors({ custom: 'Failed to parse file as JSON' });
            control.markAsTouched({ onlySelf: true });
            return;
        }
        this.onFile(event.dataTransfer.files);
        this.dragActive = false;
    };
    /**
     * @param {?} event
     * @return {?}
     */
    FormFieldFileComponent.prototype.onDragOver = /**
     * @param {?} event
     * @return {?}
     */
    function (event) {
        event.stopPropagation();
        event.preventDefault();
        this.dragActive = true;
    };
    /**
     * @return {?}
     */
    FormFieldFileComponent.prototype.onDragLeave = /**
     * @return {?}
     */
    function () {
        this.dragActive = false;
    };
    FormFieldFileComponent.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Component"], args: [{
                    selector: 'bcu-form-field-file',
                    template: "<div [formGroup]=\"form\" class=\"bcu-form-field-file\" fxFlex>\n  <mat-form-field floatLabel=\"always\" class=\"mat-form-field--no-underline\" fxFlex>\n    <mat-label *ngIf=\"field.label\">{{ field.label }}</mat-label>\n    <label\n      class=\"bcu-form-field-file-drop\"\n      [ngClass]=\"{ active: dragActive }\"\n      (drop)=\"onDrop($event)\"\n      (dragover)=\"onDragOver($event)\"\n      (dragleave)=\"onDragLeave()\"\n    >\n      <input\n        bcuMatFileInput\n        type=\"file\"\n        style=\"display:none\"\n        [accept]=\"field.accept\"\n        [id]=\"field.key\"\n        [formControlName]=\"field.key\"\n        [required]=\"field.required\"\n        [bcuDisableControl]=\"field.isDisabled(form.value)\"\n        (change)=\"onFileSelect($event)\"\n      />\n      <span>{{ field.placeholder }}</span>\n    </label>\n\n    <mat-hint *ngIf=\"field.description\" align=\"start\">{{ field.description }}</mat-hint>\n    <mat-error *ngIf=\"isTouched && !isValid\">{{ errorMessage }}</mat-error>\n  </mat-form-field>\n</div>\n",
                    encapsulation: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ViewEncapsulation"].None,
                    styles: [".bcu-form-field-file .mat-form-field--no-underline .mat-form-field-ripple,.bcu-form-field-file .mat-form-field--no-underline .mat-form-field-underline{background-color:transparent}.bcu-form-field-file .bcu-form-field-file-drop{display:block;margin:auto;height:100px;border:1px dotted rgba(0,0,0,.42);border-radius:16px;text-align:center;line-height:100px;text-transform:uppercase;cursor:pointer;font-size:16px;color:rgba(0,0,0,.54)}.bcu-form-field-file .bcu-form-field-file-drop.active{border:1px solid rgba(0,0,0,.42);background:rgba(0,0,0,.1)}"]
                }] }
    ];
    FormFieldFileComponent.propDecorators = {
        field: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }]
    };
    return FormFieldFileComponent;
}(FormFieldComponent));

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FileValueAccessorDirective = /** @class */ (function () {
    function FileValueAccessorDirective() {
        this.onChange = (/**
         * @param {?} _
         * @return {?}
         */
        function (_) { });
        this.onTouched = (/**
         * @return {?}
         */
        function () { });
    }
    /**
     * @param {?} value
     * @return {?}
     */
    FileValueAccessorDirective.prototype.writeValue = /**
     * @param {?} value
     * @return {?}
     */
    function (value) { };
    /**
     * @param {?} fn
     * @return {?}
     */
    FileValueAccessorDirective.prototype.registerOnChange = /**
     * @param {?} fn
     * @return {?}
     */
    function (fn) {
        this.onChange = fn;
    };
    /**
     * @param {?} fn
     * @return {?}
     */
    FileValueAccessorDirective.prototype.registerOnTouched = /**
     * @param {?} fn
     * @return {?}
     */
    function (fn) {
        this.onTouched = fn;
    };
    FileValueAccessorDirective.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Directive"], args: [{
                    // tslint:disable-next-line directive-selector
                    selector: 'input[type=file]',
                    providers: [{ provide: _angular_forms__WEBPACK_IMPORTED_MODULE_20__["NG_VALUE_ACCESSOR"], useExisting: FileValueAccessorDirective, multi: true }],
                },] }
    ];
    FileValueAccessorDirective.propDecorators = {
        onChange: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["HostListener"], args: ['change', ['$event.target.files'],] }],
        onTouched: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["HostListener"], args: ['blur',] }]
    };
    return FileValueAccessorDirective;
}());

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
// Boilerplate for applying mixins to MatInput.
var  
// Boilerplate for applying mixins to MatInput.
MatInputBase = /** @class */ (function () {
    function MatInputBase(_defaultErrorStateMatcher, _parentForm, _parentFormGroup, ngControl) {
        this._defaultErrorStateMatcher = _defaultErrorStateMatcher;
        this._parentForm = _parentForm;
        this._parentFormGroup = _parentFormGroup;
        this.ngControl = ngControl;
    }
    return MatInputBase;
}());
/** @type {?} */
var _MatInputMixinBase = Object(_angular_material_core__WEBPACK_IMPORTED_MODULE_13__["mixinErrorState"])(MatInputBase);
var MatFileInputDirective = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(MatFileInputDirective, _super);
    function MatFileInputDirective(_elementRef, ngControl, _parentForm, _parentFormGroup, _defaultErrorStateMatcher, inputValueAccessor) {
        var _this = _super.call(this, _defaultErrorStateMatcher, _parentForm, _parentFormGroup, ngControl) || this;
        _this._elementRef = _elementRef;
        _this.ngControl = ngControl;
        _this._uid = "bcu-mat-file-input-" + MatFileInputDirective.nextId++;
        /**
         * Implemented as part of MatFormFieldControl.
         */
        _this.focused = false;
        /**
         * Implemented as part of MatFormFieldControl.
         */
        _this.stateChanges = new rxjs__WEBPACK_IMPORTED_MODULE_15__["Subject"]();
        /**
         * Implemented as part of MatFormFieldControl.
         */
        _this.controlType = 'bcu-mat-file-input';
        /**
         * Implemented as part of MatFormFieldControl.
         */
        _this.autofilled = false;
        _this._disabled = false;
        _this._required = false;
        _this._readonly = false;
        /** @type {?} */
        var element = _this._elementRef.nativeElement;
        // If no input value accessor was explicitly specified, use the element as the input value
        // accessor.
        _this._inputValueAccessor = inputValueAccessor || element;
        _this._previousNativeValue = _this.value;
        // Force setter to be called in case id was not specified.
        _this.id = _this.id;
        return _this;
    }
    Object.defineProperty(MatFileInputDirective.prototype, "disabled", {
        /**
         * Implemented as part of MatFormFieldControl.
         */
        get: /**
         * Implemented as part of MatFormFieldControl.
         * @return {?}
         */
        function () {
            if (this.ngControl && this.ngControl.disabled !== null) {
                return this.ngControl.disabled;
            }
            return this._disabled;
        },
        set: /**
         * @param {?} value
         * @return {?}
         */
        function (value) {
            this._disabled = Object(_angular_cdk_coercion__WEBPACK_IMPORTED_MODULE_16__["coerceBooleanProperty"])(value);
            // Browsers may not fire the blur event if the input is disabled too quickly.
            // Reset from here to ensure that the element doesn't become stuck.
            if (this.focused) {
                this.focused = false;
                this.stateChanges.next();
            }
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(MatFileInputDirective.prototype, "id", {
        /**
         * Implemented as part of MatFormFieldControl.
         */
        get: /**
         * Implemented as part of MatFormFieldControl.
         * @return {?}
         */
        function () {
            return this._id;
        },
        set: /**
         * @param {?} value
         * @return {?}
         */
        function (value) {
            this._id = value || this._uid;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(MatFileInputDirective.prototype, "required", {
        /**
         * Implemented as part of MatFormFieldControl.
         */
        get: /**
         * Implemented as part of MatFormFieldControl.
         * @return {?}
         */
        function () {
            return this._required;
        },
        set: /**
         * @param {?} value
         * @return {?}
         */
        function (value) {
            this._required = Object(_angular_cdk_coercion__WEBPACK_IMPORTED_MODULE_16__["coerceBooleanProperty"])(value);
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(MatFileInputDirective.prototype, "value", {
        /**
         * Implemented as part of MatFormFieldControl.
         */
        get: /**
         * Implemented as part of MatFormFieldControl.
         * @return {?}
         */
        function () {
            return this._inputValueAccessor.value;
        },
        set: /**
         * @param {?} value
         * @return {?}
         */
        function (value) {
            if (value !== this.value) {
                this._inputValueAccessor.value = value;
                this.stateChanges.next();
            }
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(MatFileInputDirective.prototype, "readonly", {
        /** Whether the element is readonly. */
        get: /**
         * Whether the element is readonly.
         * @return {?}
         */
        function () {
            return this._readonly;
        },
        set: /**
         * @param {?} value
         * @return {?}
         */
        function (value) {
            this._readonly = Object(_angular_cdk_coercion__WEBPACK_IMPORTED_MODULE_16__["coerceBooleanProperty"])(value);
        },
        enumerable: true,
        configurable: true
    });
    /**
     * @return {?}
     */
    MatFileInputDirective.prototype.ngOnChanges = /**
     * @return {?}
     */
    function () {
        this.stateChanges.next();
    };
    /**
     * @return {?}
     */
    MatFileInputDirective.prototype.ngOnDestroy = /**
     * @return {?}
     */
    function () {
        this.stateChanges.complete();
    };
    /**
     * @return {?}
     */
    MatFileInputDirective.prototype.ngDoCheck = /**
     * @return {?}
     */
    function () {
        if (this.ngControl) {
            // We need to re-evaluate this on every change detection cycle, because there are some
            // error triggers that we can't subscribe to (e.g. parent form submissions). This means
            // that whatever logic is in here has to be super lean or we risk destroying the performance.
            this.updateErrorState();
        }
        // We need to dirty-check the native element's value, because there are some cases where
        // we won't be notified when it changes (e.g. the consumer isn't using forms or they're
        // updating the value using `emitEvent: false`).
        this._dirtyCheckNativeValue();
    };
    /** Focuses the input. */
    /**
     * Focuses the input.
     * @return {?}
     */
    MatFileInputDirective.prototype.focus = /**
     * Focuses the input.
     * @return {?}
     */
    function () {
        this._elementRef.nativeElement.focus();
    };
    /** Callback for the cases where the focused state of the input changes. */
    /**
     * Callback for the cases where the focused state of the input changes.
     * @param {?} isFocused
     * @return {?}
     */
    MatFileInputDirective.prototype._focusChanged = /**
     * Callback for the cases where the focused state of the input changes.
     * @param {?} isFocused
     * @return {?}
     */
    function (isFocused) {
        if (isFocused !== this.focused && (!this.readonly || !isFocused)) {
            this.focused = isFocused;
            this.stateChanges.next();
        }
    };
    /**
     * @return {?}
     */
    MatFileInputDirective.prototype._onInput = /**
     * @return {?}
     */
    function () {
        // This is a noop function and is used to let Angular know whenever the value changes.
        // Angular will run a new change detection each time the `input` event has been dispatched.
        // It's necessary that Angular recognizes the value change, because when floatingLabel
        // is set to false and Angular forms aren't used, the placeholder won't recognize the
        // value changes and will not disappear.
        // Listening to the input event wouldn't be necessary when the input is using the
        // FormsModule or ReactiveFormsModule, because Angular forms also listens to input events.
    };
    /** Does some manual dirty checking on the native input `value` property. */
    /**
     * Does some manual dirty checking on the native input `value` property.
     * @protected
     * @return {?}
     */
    MatFileInputDirective.prototype._dirtyCheckNativeValue = /**
     * Does some manual dirty checking on the native input `value` property.
     * @protected
     * @return {?}
     */
    function () {
        /** @type {?} */
        var newValue = this._elementRef.nativeElement.value;
        if (this._previousNativeValue !== newValue) {
            this._previousNativeValue = newValue;
            this.stateChanges.next();
        }
    };
    Object.defineProperty(MatFileInputDirective.prototype, "empty", {
        /**
         * Implemented as part of MatFormFieldControl.
         */
        get: /**
         * Implemented as part of MatFormFieldControl.
         * @return {?}
         */
        function () {
            return false;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(MatFileInputDirective.prototype, "shouldLabelFloat", {
        /**
         * Implemented as part of MatFormFieldControl.
         */
        get: /**
         * Implemented as part of MatFormFieldControl.
         * @return {?}
         */
        function () {
            return true;
        },
        enumerable: true,
        configurable: true
    });
    /**
     * Implemented as part of MatFormFieldControl.
     */
    /**
     * Implemented as part of MatFormFieldControl.
     * @param {?} ids
     * @return {?}
     */
    MatFileInputDirective.prototype.setDescribedByIds = /**
     * Implemented as part of MatFormFieldControl.
     * @param {?} ids
     * @return {?}
     */
    function (ids) {
        this._ariaDescribedby = ids.join(' ');
    };
    /**
     * Implemented as part of MatFormFieldControl.
     */
    /**
     * Implemented as part of MatFormFieldControl.
     * @return {?}
     */
    MatFileInputDirective.prototype.onContainerClick = /**
     * Implemented as part of MatFormFieldControl.
     * @return {?}
     */
    function () {
        // Do not re-focus the input element if the element is already focused.
        if (!this.focused) {
            this.focus();
        }
    };
    MatFileInputDirective.nextId = 0;
    MatFileInputDirective.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Directive"], args: [{
                    selector: 'input[type=file][bcuMatFileInput]',
                    exportAs: 'bcuMatFileInput',
                    providers: [{ provide: _angular_material__WEBPACK_IMPORTED_MODULE_14__["MatFormFieldControl"], useExisting: MatFileInputDirective }],
                    // tslint:disable-next-line
                    host: {
                        class: 'mat-input-element',
                        // Native input properties that are overwritten by Angular inputs need to be synced with
                        // the native input element. Otherwise property bindings for those don't work.
                        '[attr.id]': 'id',
                        '[disabled]': 'disabled',
                        '[required]': 'required',
                        '[attr.readonly]': 'readonly',
                        '[accept]': 'accept',
                        '[attr.aria-describedby]': '_ariaDescribedby || null',
                        '[attr.aria-invalid]': 'errorState',
                        '[attr.aria-required]': 'required.toString()',
                        '(blur)': '_focusChanged(false)',
                        '(focus)': '_focusChanged(true)',
                        '(input)': '_onInput()',
                    },
                },] }
    ];
    /** @nocollapse */
    MatFileInputDirective.ctorParameters = function () { return [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["ElementRef"] },
        { type: _angular_forms__WEBPACK_IMPORTED_MODULE_20__["NgControl"], decorators: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Optional"] }, { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Self"] }] },
        { type: _angular_forms__WEBPACK_IMPORTED_MODULE_20__["NgForm"], decorators: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Optional"] }] },
        { type: _angular_forms__WEBPACK_IMPORTED_MODULE_20__["FormGroupDirective"], decorators: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Optional"] }] },
        { type: _angular_material_core__WEBPACK_IMPORTED_MODULE_13__["ErrorStateMatcher"] },
        { type: undefined, decorators: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Optional"] }, { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Self"] }, { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Inject"], args: [_angular_material__WEBPACK_IMPORTED_MODULE_14__["MAT_INPUT_VALUE_ACCESSOR"],] }] }
    ]; };
    MatFileInputDirective.propDecorators = {
        accept: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        disabled: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        id: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        placeholder: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        required: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        errorStateMatcher: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        value: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }],
        readonly: [{ type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Input"] }]
    };
    return MatFileInputDirective;
}(_MatInputMixinBase));

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var UniformModule = /** @class */ (function () {
    function UniformModule() {
    }
    UniformModule.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["NgModule"], args: [{
                    imports: [
                        _angular_common__WEBPACK_IMPORTED_MODULE_0__["CommonModule"],
                        _angular_forms__WEBPACK_IMPORTED_MODULE_20__["FormsModule"],
                        _angular_forms__WEBPACK_IMPORTED_MODULE_20__["ReactiveFormsModule"],
                        _angular_platform_browser_animations__WEBPACK_IMPORTED_MODULE_1__["BrowserAnimationsModule"],
                        _angular_material_input__WEBPACK_IMPORTED_MODULE_2__["MatInputModule"],
                        _angular_material_checkbox__WEBPACK_IMPORTED_MODULE_3__["MatCheckboxModule"],
                        _angular_material_select__WEBPACK_IMPORTED_MODULE_4__["MatSelectModule"],
                        _angular_material_button__WEBPACK_IMPORTED_MODULE_5__["MatButtonModule"],
                        _angular_flex_layout__WEBPACK_IMPORTED_MODULE_6__["FlexLayoutModule"],
                        _ctrl_ngx_codemirror__WEBPACK_IMPORTED_MODULE_7__["CodemirrorModule"],
                    ],
                    exports: [
                        FormFieldCodeComponent,
                        FormFieldInputComponent,
                        FormFieldTextareaComponent,
                        FormFieldSelectComponent,
                        FormFieldCheckboxComponent,
                        FormFieldFileComponent,
                        FormFieldComponent,
                        FormComponent,
                        DisableControlDirective,
                        FormGroupComponent,
                        FormFieldContainerComponent,
                        FormFieldCodeComponent,
                        FormFieldInputComponent,
                        FormFieldTextareaComponent,
                        FormFieldSelectComponent,
                        FormFieldCheckboxComponent,
                        FormFieldFileComponent,
                        FileValueAccessorDirective,
                        MatFileInputDirective,
                    ],
                    entryComponents: [
                        FormFieldCodeComponent,
                        FormFieldInputComponent,
                        FormFieldTextareaComponent,
                        FormFieldSelectComponent,
                        FormFieldCheckboxComponent,
                        FormFieldFileComponent,
                    ],
                    declarations: [
                        FormFieldComponent,
                        FormComponent,
                        DisableControlDirective,
                        FormGroupComponent,
                        FormFieldContainerComponent,
                        FormFieldCodeComponent,
                        FormFieldInputComponent,
                        FormFieldTextareaComponent,
                        FormFieldSelectComponent,
                        FormFieldCheckboxComponent,
                        FormFieldFileComponent,
                        FileValueAccessorDirective,
                        MatFileInputDirective,
                    ],
                },] }
    ];
    return UniformModule;
}());

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormField = /** @class */ (function () {
    function FormField(_a) {
        var controlType = _a.controlType, key = _a.key, value = _a.value, label = _a.label, _b = _a.required, required = _b === void 0 ? false : _b, _c = _a.hidden, hidden = _c === void 0 ? false : _c, _d = _a.disabled, disabled = _d === void 0 ? false : _d, _e = _a.placeholder, placeholder = _e === void 0 ? '' : _e, _f = _a.description, description = _f === void 0 ? '' : _f, minLength = _a.minLength, maxLength = _a.maxLength, pattern = _a.pattern, showIf = _a.showIf, _g = _a.validators, validators = _g === void 0 ? [] : _g, _h = _a.default, defaultValue = _h === void 0 ? null : _h;
        this.CONTROL_TYPES = CONTROL_TYPES;
        this.key = key;
        if (!this.key) {
            throw new TypeError('key is required');
        }
        this.controlType = controlType;
        if (CONTROL_TYPES[(/** @type {?} */ (controlType))]) {
            this.controlType = CONTROL_TYPES[(/** @type {?} */ (controlType))];
        }
        if (!this.controlType) {
            throw new TypeError('controlType is required');
        }
        this.label = label;
        if (typeof this.label === 'undefined') {
            throw new TypeError('label is required');
        }
        this.validators = validators;
        this.required = required;
        if (required) {
            this.validators.push(_angular_forms__WEBPACK_IMPORTED_MODULE_20__["Validators"].required);
        }
        this.minLength = minLength;
        if (minLength) {
            this.validators.push(_angular_forms__WEBPACK_IMPORTED_MODULE_20__["Validators"].minLength(minLength));
        }
        this.maxLength = maxLength;
        if (maxLength) {
            this.validators.push(_angular_forms__WEBPACK_IMPORTED_MODULE_20__["Validators"].maxLength(maxLength));
        }
        this.pattern = pattern;
        if (pattern) {
            this.validators.push(_angular_forms__WEBPACK_IMPORTED_MODULE_20__["Validators"].pattern(pattern));
        }
        if (showIf) {
            /** @type {?} */
            var schema = Object.assign({}, showIf, {
                additionalProperties: true,
            });
            /** @type {?} */
            var ajv = new ajv__WEBPACK_IMPORTED_MODULE_19__({ validateSchema: 'log', missingRefs: 'ignore' });
            this.showIfSchema = ajv.compile(schema);
        }
        this.hidden = hidden;
        this.disabled = disabled;
        this.default = value != null ? value : defaultValue;
        this.description = description;
        this.placeholder = placeholder;
    }
    Object.defineProperty(FormField.prototype, "control", {
        get: /**
         * @return {?}
         */
        function () {
            return new _angular_forms__WEBPACK_IMPORTED_MODULE_20__["FormControl"]({
                value: this.default,
                disabled: this.disabled,
            }, this.validators);
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(FormField.prototype, "component", {
        get: /**
         * @return {?}
         */
        function () {
            return null;
        },
        enumerable: true,
        configurable: true
    });
    /**
     * @param {?} form
     * @return {?}
     */
    FormField.prototype.isHidden = /**
     * @param {?} form
     * @return {?}
     */
    function (form) {
        return this.hidden || (this.showIfSchema && !this.showIfSchema(form));
    };
    /**
     * @param {?} form
     * @return {?}
     */
    FormField.prototype.isDisabled = /**
     * @param {?} form
     * @return {?}
     */
    function (form) {
        return this.disabled === true || (this.showIfSchema && !this.showIfSchema(form));
    };
    return FormField;
}());

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldCheckbox = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldCheckbox, _super);
    function FormFieldCheckbox(field) {
        return _super.call(this, Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__assign"])({}, field, { controlType: CONTROL_TYPES.CHECKBOX })) || this;
    }
    Object.defineProperty(FormFieldCheckbox.prototype, "component", {
        get: /**
         * @return {?}
         */
        function () {
            return FormFieldCheckboxComponent;
        },
        enumerable: true,
        configurable: true
    });
    return FormFieldCheckbox;
}(FormField));

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldCode = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldCode, _super);
    function FormFieldCode(_a) {
        var _b = _a.options, options = _b === void 0 ? {} : _b, field = Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__rest"])(_a, ["options"]);
        var _this = _super.call(this, Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__assign"])({}, field, { controlType: CONTROL_TYPES.CODE })) || this;
        _this.options = options;
        return _this;
    }
    Object.defineProperty(FormFieldCode.prototype, "component", {
        get: /**
         * @return {?}
         */
        function () {
            return FormFieldCodeComponent;
        },
        enumerable: true,
        configurable: true
    });
    return FormFieldCode;
}(FormField));

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldFile = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldFile, _super);
    function FormFieldFile(_a) {
        var _b = _a.fillForm, fillForm = _b === void 0 ? true : _b, _c = _a.accept, accept = _c === void 0 ? '.json,application/json' : _c, field = Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__rest"])(_a, ["fillForm", "accept"]);
        var _this = _super.call(this, Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__assign"])({ placeholder: 'Click here to select or drop a file' }, field, { controlType: CONTROL_TYPES.FILE })) || this;
        _this.fillForm = fillForm;
        _this.accept = accept;
        return _this;
    }
    Object.defineProperty(FormFieldFile.prototype, "component", {
        get: /**
         * @return {?}
         */
        function () {
            return FormFieldFileComponent;
        },
        enumerable: true,
        configurable: true
    });
    return FormFieldFile;
}(FormField));

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldText = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldText, _super);
    function FormFieldText(field) {
        return _super.call(this, Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__assign"])({}, field, { controlType: CONTROL_TYPES.TEXT })) || this;
    }
    Object.defineProperty(FormFieldText.prototype, "component", {
        get: /**
         * @return {?}
         */
        function () {
            return FormFieldInputComponent;
        },
        enumerable: true,
        configurable: true
    });
    return FormFieldText;
}(FormField));
var FormFieldNumber = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldNumber, _super);
    function FormFieldNumber(_a) {
        var min = _a.min, max = _a.max, _b = _a.validators, validators = _b === void 0 ? [] : _b, field = Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__rest"])(_a, ["min", "max", "validators"]);
        var _this = this;
        if (min) {
            validators.push(_angular_forms__WEBPACK_IMPORTED_MODULE_20__["Validators"].min(min));
        }
        if (max) {
            validators.push(_angular_forms__WEBPACK_IMPORTED_MODULE_20__["Validators"].max(max));
        }
        _this = _super.call(this, Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__assign"])({}, field, { validators: validators, controlType: CONTROL_TYPES.NUMBER })) || this;
        _this.min = min;
        _this.max = max;
        return _this;
    }
    Object.defineProperty(FormFieldNumber.prototype, "component", {
        get: /**
         * @return {?}
         */
        function () {
            return FormFieldInputComponent;
        },
        enumerable: true,
        configurable: true
    });
    return FormFieldNumber;
}(FormField));
var FormFieldPassword = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldPassword, _super);
    function FormFieldPassword(field) {
        return _super.call(this, Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__assign"])({}, field, { controlType: CONTROL_TYPES.PASSWORD })) || this;
    }
    Object.defineProperty(FormFieldPassword.prototype, "component", {
        get: /**
         * @return {?}
         */
        function () {
            return FormFieldInputComponent;
        },
        enumerable: true,
        configurable: true
    });
    return FormFieldPassword;
}(FormField));

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldSelect = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldSelect, _super);
    function FormFieldSelect(_a) {
        var _b = _a.options, options = _b === void 0 ? [] : _b, field = Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__rest"])(_a, ["options"]);
        var _this = _super.call(this, Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__assign"])({}, field, { controlType: CONTROL_TYPES.SELECT })) || this;
        _this.options = [];
        _this.options = options;
        return _this;
    }
    Object.defineProperty(FormFieldSelect.prototype, "component", {
        get: /**
         * @return {?}
         */
        function () {
            return FormFieldSelectComponent;
        },
        enumerable: true,
        configurable: true
    });
    return FormFieldSelect;
}(FormField));

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */
var FormFieldTextarea = /** @class */ (function (_super) {
    Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__extends"])(FormFieldTextarea, _super);
    function FormFieldTextarea(field) {
        return _super.call(this, Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__assign"])({}, field, { controlType: CONTROL_TYPES.TEXTAREA })) || this;
    }
    Object.defineProperty(FormFieldTextarea.prototype, "component", {
        get: /**
         * @return {?}
         */
        function () {
            return FormFieldTextareaComponent;
        },
        enumerable: true,
        configurable: true
    });
    return FormFieldTextarea;
}(FormField));

var _a, _b;
var UniformService = /** @class */ (function () {
    function UniformService() {
    }
    /**
     * @param {?=} groups
     * @return {?}
     */
    UniformService.factory = /**
     * @param {?=} groups
     * @return {?}
     */
    function (groups) {
        if (groups === void 0) { groups = []; }
        return groups.map((/**
         * @param {?} group
         * @return {?}
         */
        function (group) { return (Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__assign"])({}, group, { fields: Object(lodash_es__WEBPACK_IMPORTED_MODULE_18__["flatten"])(group.fields.map((/**
             * @param {?} field
             * @return {?}
             */
            function (field) { return UniformService.fieldFactory(field); }))) })); }));
    };
    /**
     * @param {?} __0
     * @param {?=} throwError
     * @return {?}
     */
    UniformService.fieldFactory = /**
     * @param {?} __0
     * @param {?=} throwError
     * @return {?}
     */
    function (_a, throwError) {
        if (throwError === void 0) { throwError = false; }
        var controlType = _a.controlType, controlGroupType = _a.controlGroupType, field = Object(tslib__WEBPACK_IMPORTED_MODULE_21__["__rest"])(_a, ["controlType", "controlGroupType"]);
        /** @type {?} */
        var err;
        if (controlGroupType) {
            /** @type {?} */
            var controlGroupFieldsFactory = UniformService.CONTROL_GROUPS[controlGroupType];
            if (!controlGroupFieldsFactory) {
                err = new TypeError("Group type \"" + controlGroupType + "\" is not supported");
                if (throwError) {
                    throw err;
                }
                console.error('Ignoring field', field, err);
                return [];
            }
            return controlGroupFieldsFactory(field);
        }
        if (controlType) {
            /** @type {?} */
            var Control = UniformService.CONTROLS[controlType];
            if (!Control) {
                err = new TypeError("Type \"" + controlType + "\" is not supported");
                if (throwError) {
                    throw err;
                }
                console.error('Ignoring field', field, err);
                return [];
            }
            try {
                return [new Control(field)];
            }
            catch (err) {
                if (throwError) {
                    throw err;
                }
                console.error('Ignoring field', field, err);
                return [];
            }
        }
        err = new TypeError('Type is not specified');
        if (throwError) {
            throw err;
        }
        console.error('Ignoring field', { controlType: controlType, controlGroupType: controlGroupType, field: field }, err);
        return [];
    };
    /**
     * @param {?} __0
     * @return {?}
     */
    UniformService.amazonFields = /**
     * @param {?} __0
     * @return {?}
     */
    function (_a) {
        var key = _a.key, showIf = _a.showIf, _b = _a.value, value = _b === void 0 ? {} : _b;
        return [
            new FormFieldText({
                key: key + ".AWS_ACCESS_KEY_ID",
                label: 'AWS Access Key ID',
                required: true,
                value: value['AWS_ACCESS_KEY_ID'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".AWS_SECRET_ACCESS_KEY",
                label: 'AWS Secret Access Key',
                required: true,
                value: value['AWS_SECRET_ACCESS_KEY'],
                showIf: showIf,
            }),
        ];
    };
    /**
     * @param {?} __0
     * @return {?}
     */
    UniformService.alibabaFields = /**
     * @param {?} __0
     * @return {?}
     */
    function (_a) {
        var key = _a.key, showIf = _a.showIf, _b = _a.value, value = _b === void 0 ? {} : _b;
        return [
            new FormFieldText({
                key: key + ".ALIBABA_ACCESS_KEY_ID",
                label: 'Alibaba Access Key ID',
                required: true,
                value: value['ALIBABA_ACCESS_KEY_ID'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".ALIBABA_ACCESS_KEY_SECRET",
                label: 'Alibaba Secret Access Key',
                required: true,
                value: value['ALIBABA_ACCESS_KEY_SECRET'],
                showIf: showIf,
            }),
        ];
    };
    /**
     * @param {?} __0
     * @return {?}
     */
    UniformService.azureFields = /**
     * @param {?} __0
     * @return {?}
     */
    function (_a) {
        var key = _a.key, showIf = _a.showIf, _b = _a.value, value = _b === void 0 ? {} : _b;
        return [
            new FormFieldText({
                key: key + ".AZURE_CLIENT_ID",
                label: 'Azure Client ID',
                required: true,
                value: value['AZURE_CLIENT_ID'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".AZURE_CLIENT_SECRET",
                label: 'Azure Client Secret',
                required: true,
                value: value['AZURE_CLIENT_SECRET'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".AZURE_TENANT_ID",
                label: 'Azure Tenant ID',
                required: true,
                value: value['AZURE_TENANT_ID'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".AZURE_SUBSCRIPTION_ID",
                label: 'Azure Subscription ID',
                required: true,
                value: value['AZURE_SUBSCRIPTION_ID'],
                showIf: showIf,
            }),
        ];
    };
    /**
     * @param {?} __0
     * @return {?}
     */
    UniformService.googleFields = /**
     * @param {?} __0
     * @return {?}
     */
    function (_a) {
        var key = _a.key, showIf = _a.showIf, _b = _a.value, value = _b === void 0 ? {} : _b;
        return [
            new FormFieldFile({
                key: key + ".json_key",
                label: 'Service account key',
                description: 'Fill out the form by providing a JSON key file',
                value: value['json_key'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".service_account",
                label: 'Type',
                required: true,
                default: 'service_account',
                value: value['service_account'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".project_id",
                label: 'Project Id',
                required: true,
                value: value['project_id'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".private_key_id",
                label: 'Project Key Id',
                required: true,
                value: value['private_key_id'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".private_key",
                label: 'Private Key',
                required: true,
                value: value['private_key'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".client_email",
                label: 'Client email',
                required: true,
                value: value['client_email'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".client_id",
                label: 'Client Id',
                required: true,
                value: value['client_id'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".auth_uri",
                label: 'Auth URI',
                required: true,
                default: 'https://accounts.google.com/o/oauth2/auth',
                value: value['auth_uri'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".token_uri",
                label: 'Token URI',
                required: true,
                default: 'https://accounts.google.com/o/oauth2/token',
                value: value['token_uri'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".auth_provider_x509_cert_url",
                label: 'Auth Provider X509 Cert URL',
                required: true,
                default: 'https://www.googleapis.com/oauth2/v1/certs',
                value: value['auth_provider_x509_cert_url'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".client_x509_cert_url",
                label: 'Client X509 Cert URL',
                required: true,
                value: value['client_x509_cert_url'],
                showIf: showIf,
            }),
        ];
    };
    /**
     * @param {?} __0
     * @return {?}
     */
    UniformService.oracleFields = /**
     * @param {?} __0
     * @return {?}
     */
    function (_a) {
        var key = _a.key, showIf = _a.showIf, _b = _a.value, value = _b === void 0 ? {} : _b;
        return [
            new FormFieldText({
                key: key + ".user_ocid",
                label: 'User OCID',
                required: true,
                value: value['user_ocid'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".api_key",
                label: 'API key',
                required: true,
                value: value['api_key'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".api_key_fingerprint",
                label: 'API key fingerprint',
                required: true,
                value: value['api_key_fingerprint'],
                showIf: showIf,
            }),
            new FormFieldSelect({
                key: key + ".region",
                label: 'Region',
                required: true,
                value: value['region'],
                options: [
                    {
                        label: 'EU West (Frankfurt, Germany)',
                        value: 'eu-frankfurt-1',
                    },
                    {
                        label: 'EU West (London, United Kingdom)',
                        value: 'uk-london-1',
                    },
                    {
                        label: 'US East (Ashburn, VA)',
                        value: 'us-ashburn-1',
                    },
                    {
                        label: 'US West (Phoenix, AZ)',
                        value: 'us-phoenix-1',
                    },
                ],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".compartment_ocid",
                label: 'Compartment OCID',
                required: true,
                value: value['compartment_ocid'],
                showIf: showIf,
            }),
            new FormFieldText({
                key: key + ".tenancy_ocid",
                label: 'Tenancy OCID',
                required: true,
                value: value['tenancy_ocid'],
                showIf: showIf,
            }),
        ];
    };
    /**
     * @param {?} __0
     * @return {?}
     */
    UniformService.passwordFields = /**
     * @param {?} __0
     * @return {?}
     */
    function (_a) {
        var key = _a.key, showIf = _a.showIf, _b = _a.value, value = _b === void 0 ? {} : _b;
        return [
            new FormFieldText({
                key: key + ".username",
                label: 'username',
                required: true,
                value: value['username'],
                showIf: showIf,
            }),
            new FormFieldPassword({
                key: key + ".password",
                label: 'password',
                minLength: 8,
                placeholder: 'Auto-generated when left empty',
                value: value['password'],
                showIf: showIf,
            }),
        ];
    };
    /**
     * @param {?} __0
     * @return {?}
     */
    UniformService.tlsFields = /**
     * @param {?} __0
     * @return {?}
     */
    function (_a) {
        var key = _a.key, showIf = _a.showIf, _b = _a.value, value = _b === void 0 ? {} : _b;
        return [
            new FormFieldText({
                key: key + ".hosts",
                label: 'Hosts',
                description: 'Host names separated by commas',
                placeholder: 'example.com',
                pattern: '^(?:(?:^|,))((([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]).)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9-]*[A-Za-z0-9]))$',
                value: value['hosts'],
                showIf: showIf,
            }),
            new FormFieldPassword({
                key: key + ".caCert",
                placeholder: 'ca.cert',
                label: 'Certificate authority certificate',
                value: value['caCert'],
                showIf: showIf,
            }),
            new FormFieldPassword({
                key: key + ".caKey",
                placeholder: 'ca.key',
                label: 'Certificate authority key',
                value: value['caKey'],
                showIf: showIf,
            }),
            new FormFieldPassword({
                key: key + ".serverCert",
                placeholder: 'server.key',
                label: 'Server certificate',
                value: value['serverCert'],
                showIf: showIf,
            }),
            new FormFieldPassword({
                key: key + ".serverKey",
                placeholder: 'server.key',
                label: 'Server key',
                value: value['serverKey'],
                showIf: showIf,
            }),
            new FormFieldPassword({
                key: key + ".clientCert",
                placeholder: 'client.key',
                label: 'Client certificate',
                value: value['clientCert'],
                showIf: showIf,
            }),
            new FormFieldPassword({
                key: key + ".clientKey",
                placeholder: 'client.key',
                label: 'Client key',
                value: value['clientKey'],
                showIf: showIf,
            }),
        ];
    };
    UniformService.CONTROLS = (_a = {},
        _a[CONTROL_TYPES.CHECKBOX] = FormFieldCheckbox,
        _a[CONTROL_TYPES.CODE] = FormFieldCode,
        _a[CONTROL_TYPES.FILE] = FormFieldFile,
        _a[CONTROL_TYPES.NUMBER] = FormFieldNumber,
        _a[CONTROL_TYPES.PASSWORD] = FormFieldPassword,
        _a[CONTROL_TYPES.SELECT] = FormFieldSelect,
        _a[CONTROL_TYPES.TEXT] = FormFieldText,
        _a[CONTROL_TYPES.TEXTAREA] = FormFieldTextarea,
        _a);
    UniformService.CONTROL_GROUPS = (_b = {},
        _b[CONTROL_GROUP_TYPES.AMAZON] = UniformService.amazonFields,
        _b[CONTROL_GROUP_TYPES.AZURE] = UniformService.azureFields,
        _b[CONTROL_GROUP_TYPES.ALIBABA] = UniformService.alibabaFields,
        _b[CONTROL_GROUP_TYPES.GOOGLE] = UniformService.googleFields,
        _b[CONTROL_GROUP_TYPES.ORACLE] = UniformService.oracleFields,
        _b[CONTROL_GROUP_TYPES.PASSWORD] = UniformService.passwordFields,
        _b[CONTROL_GROUP_TYPES.TLS] = UniformService.tlsFields,
        _b);
    UniformService.decorators = [
        { type: _angular_core__WEBPACK_IMPORTED_MODULE_17__["Injectable"], args: [{
                    providedIn: 'root',
                },] }
    ];
    /** @nocollapse */ UniformService.ngInjectableDef = Object(_angular_core__WEBPACK_IMPORTED_MODULE_17__["defineInjectable"])({ factory: function UniformService_Factory() { return new UniformService(); }, token: UniformService, providedIn: "root" });
    return UniformService;
}());

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */

/**
 * @fileoverview added by tsickle
 * @suppress {checkTypes,extraRequire,missingOverride,missingReturn,unusedPrivateMembers,uselessCode} checked by tsc
 */



//# sourceMappingURL=banzaicloud-uniform.js.map

/***/ }),

/***/ "./src/$$_lazy_route_resource lazy recursive":
/*!**********************************************************!*\
  !*** ./src/$$_lazy_route_resource lazy namespace object ***!
  \**********************************************************/
/*! no static exports found */
/***/ (function(module, exports) {

function webpackEmptyAsyncContext(req) {
	// Here Promise.resolve().then() is used instead of new Promise() to prevent
	// uncaught exception popping up in devtools
	return Promise.resolve().then(function() {
		var e = new Error("Cannot find module '" + req + "'");
		e.code = 'MODULE_NOT_FOUND';
		throw e;
	});
}
webpackEmptyAsyncContext.keys = function() { return []; };
webpackEmptyAsyncContext.resolve = webpackEmptyAsyncContext;
module.exports = webpackEmptyAsyncContext;
webpackEmptyAsyncContext.id = "./src/$$_lazy_route_resource lazy recursive";

/***/ }),

/***/ "./src/app/app.component.html":
/*!************************************!*\
  !*** ./src/app/app.component.html ***!
  \************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = "<div>\n  <a style=\"display: none;\" target=\"_blank\" rel=\"noopener noreferrer\" download=\"anwsers.yaml\" #downloadLink></a>\n  <mat-card class=\"container\">\n    <bcu-form\n      [groups]=\"groups | async\"\n      [initialValue]=\"initialValue\"\n      (valueChanges)=\"onValueChanges($event)\"\n      (save)=\"onSave($event)\"\n    ></bcu-form>\n  </mat-card>\n</div>\n"

/***/ }),

/***/ "./src/app/app.component.scss":
/*!************************************!*\
  !*** ./src/app/app.component.scss ***!
  \************************************/
/*! no static exports found */
/***/ (function(module, exports) {

module.exports = ".container {\n  max-width: 960px;\n  padding: 24px;\n  margin: 0 auto; }\n\n/*# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJzb3VyY2VzIjpbIi9Vc2Vycy9hbmRyYXN0b3RoL0RldmVsb3Blci9zcmMvZ2l0aHViLmNvbS9iYW56YWljbG91ZC91bmlmb3JtL3NyYy9hcHAvYXBwLmNvbXBvbmVudC5zY3NzIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQUFBO0VBQ0UsZ0JBQWdCO0VBQ2hCLGFBQWE7RUFDYixjQUFjLEVBQUEiLCJmaWxlIjoic3JjL2FwcC9hcHAuY29tcG9uZW50LnNjc3MiLCJzb3VyY2VzQ29udGVudCI6WyIuY29udGFpbmVyIHtcbiAgbWF4LXdpZHRoOiA5NjBweDtcbiAgcGFkZGluZzogMjRweDtcbiAgbWFyZ2luOiAwIGF1dG87XG59XG4iXX0= */"

/***/ }),

/***/ "./src/app/app.component.ts":
/*!**********************************!*\
  !*** ./src/app/app.component.ts ***!
  \**********************************/
/*! exports provided: AppComponent */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "AppComponent", function() { return AppComponent; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _banzaicloud_uniform__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! @banzaicloud/uniform */ "./dist/banzaicloud/uniform/fesm5/banzaicloud-uniform.js");
/* harmony import */ var _angular_common_http__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! @angular/common/http */ "./node_modules/@angular/common/fesm5/http.js");
/* harmony import */ var rxjs_operators__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! rxjs/operators */ "./node_modules/rxjs/_esm5/operators/index.js");
/* harmony import */ var lodash_fp__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! lodash/fp */ "./node_modules/lodash/fp.js");
/* harmony import */ var lodash_fp__WEBPACK_IMPORTED_MODULE_5___default = /*#__PURE__*/__webpack_require__.n(lodash_fp__WEBPACK_IMPORTED_MODULE_5__);
/* harmony import */ var codemirror_mode_javascript_javascript__WEBPACK_IMPORTED_MODULE_6__ = __webpack_require__(/*! codemirror/mode/javascript/javascript */ "./node_modules/codemirror/mode/javascript/javascript.js");
/* harmony import */ var codemirror_mode_javascript_javascript__WEBPACK_IMPORTED_MODULE_6___default = /*#__PURE__*/__webpack_require__.n(codemirror_mode_javascript_javascript__WEBPACK_IMPORTED_MODULE_6__);
/* harmony import */ var codemirror_mode_markdown_markdown__WEBPACK_IMPORTED_MODULE_7__ = __webpack_require__(/*! codemirror/mode/markdown/markdown */ "./node_modules/codemirror/mode/markdown/markdown.js");
/* harmony import */ var codemirror_mode_markdown_markdown__WEBPACK_IMPORTED_MODULE_7___default = /*#__PURE__*/__webpack_require__.n(codemirror_mode_markdown_markdown__WEBPACK_IMPORTED_MODULE_7__);
/* harmony import */ var codemirror_mode_yaml_yaml__WEBPACK_IMPORTED_MODULE_8__ = __webpack_require__(/*! codemirror/mode/yaml/yaml */ "./node_modules/codemirror/mode/yaml/yaml.js");
/* harmony import */ var codemirror_mode_yaml_yaml__WEBPACK_IMPORTED_MODULE_8___default = /*#__PURE__*/__webpack_require__.n(codemirror_mode_yaml_yaml__WEBPACK_IMPORTED_MODULE_8__);
/* harmony import */ var codemirror_addon_display_placeholder__WEBPACK_IMPORTED_MODULE_9__ = __webpack_require__(/*! codemirror/addon/display/placeholder */ "./node_modules/codemirror/addon/display/placeholder.js");
/* harmony import */ var codemirror_addon_display_placeholder__WEBPACK_IMPORTED_MODULE_9___default = /*#__PURE__*/__webpack_require__.n(codemirror_addon_display_placeholder__WEBPACK_IMPORTED_MODULE_9__);






var reduceObject = lodash_fp__WEBPACK_IMPORTED_MODULE_5__["reduce"].convert({ cap: false });
var mapObject = lodash_fp__WEBPACK_IMPORTED_MODULE_5__["map"].convert({ cap: false });
// see https://codemirror.net/mode/index.html



// see https://codemirror.net/demo/placeholder.html

var AppComponent = /** @class */ (function () {
    function AppComponent(http) {
        this.http = http;
        this.groups = this.http
            .get('/api/v1/form')
            .pipe(Object(rxjs_operators__WEBPACK_IMPORTED_MODULE_4__["map"])(function (g) { return _banzaicloud_uniform__WEBPACK_IMPORTED_MODULE_2__["UniformService"].factory(g); }));
    }
    AppComponent.prototype.onValueChanges = function (event) {
        this.values = event;
    };
    AppComponent.prototype.onSave = function (event) {
        // NOTE download as file
        // const groups = this.rawGroups.map((group) => ({
        //   ...group,
        //   fields: group.fields.map((field) => ({
        //     ...field,
        //     value: Object.keys(event)
        //       .filter((key) => key === field.key || key.startsWith(`${field.key}.`))
        //       .reduce((value, key) => fp.set(key, event[key], value), {}),
        //   })),
        // }));
        // const data = yaml.safeDump(groups);
        // const blob = new Blob([data], { type: 'text/vnd.yaml' });
        // const url = window.URL.createObjectURL(blob);
        // this.downloadLink.nativeElement.href = url;
        // this.downloadLink.nativeElement.click();
        // window.URL.revokeObjectURL(blob);
        // this.downloadLink.nativeElement.href = '';
        var values = reduceObject(function (v, val, key) { return lodash_fp__WEBPACK_IMPORTED_MODULE_5__["set"](key, val, v); }, {})(event);
        this.http
            .post('/api/v1/form', values, {
            headers: new _angular_common_http__WEBPACK_IMPORTED_MODULE_3__["HttpHeaders"]({
                'Content-Type': 'application/json',
            }),
        })
            .subscribe();
    };
    tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["ViewChild"])('downloadLink'),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:type", _angular_core__WEBPACK_IMPORTED_MODULE_1__["ElementRef"])
    ], AppComponent.prototype, "downloadLink", void 0);
    AppComponent = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_1__["Component"])({
            selector: 'banzai-root',
            template: __webpack_require__(/*! ./app.component.html */ "./src/app/app.component.html"),
            styles: [__webpack_require__(/*! ./app.component.scss */ "./src/app/app.component.scss")]
        }),
        tslib__WEBPACK_IMPORTED_MODULE_0__["__metadata"]("design:paramtypes", [_angular_common_http__WEBPACK_IMPORTED_MODULE_3__["HttpClient"]])
    ], AppComponent);
    return AppComponent;
}());



/***/ }),

/***/ "./src/app/app.module.ts":
/*!*******************************!*\
  !*** ./src/app/app.module.ts ***!
  \*******************************/
/*! exports provided: AppModule */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "AppModule", function() { return AppModule; });
/* harmony import */ var tslib__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! tslib */ "./node_modules/tslib/tslib.es6.js");
/* harmony import */ var _angular_platform_browser__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/platform-browser */ "./node_modules/@angular/platform-browser/fesm5/platform-browser.js");
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _angular_material_card__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! @angular/material/card */ "./node_modules/@angular/material/esm5/card.es5.js");
/* harmony import */ var _angular_common_http__WEBPACK_IMPORTED_MODULE_4__ = __webpack_require__(/*! @angular/common/http */ "./node_modules/@angular/common/fesm5/http.js");
/* harmony import */ var _banzaicloud_uniform__WEBPACK_IMPORTED_MODULE_5__ = __webpack_require__(/*! @banzaicloud/uniform */ "./dist/banzaicloud/uniform/fesm5/banzaicloud-uniform.js");
/* harmony import */ var _app_component__WEBPACK_IMPORTED_MODULE_6__ = __webpack_require__(/*! ./app.component */ "./src/app/app.component.ts");







var AppModule = /** @class */ (function () {
    function AppModule() {
    }
    AppModule = tslib__WEBPACK_IMPORTED_MODULE_0__["__decorate"]([
        Object(_angular_core__WEBPACK_IMPORTED_MODULE_2__["NgModule"])({
            declarations: [_app_component__WEBPACK_IMPORTED_MODULE_6__["AppComponent"]],
            imports: [_angular_platform_browser__WEBPACK_IMPORTED_MODULE_1__["BrowserModule"], _angular_common_http__WEBPACK_IMPORTED_MODULE_4__["HttpClientModule"], _banzaicloud_uniform__WEBPACK_IMPORTED_MODULE_5__["UniformModule"], _angular_material_card__WEBPACK_IMPORTED_MODULE_3__["MatCardModule"]],
            providers: [],
            bootstrap: [_app_component__WEBPACK_IMPORTED_MODULE_6__["AppComponent"]],
        })
    ], AppModule);
    return AppModule;
}());



/***/ }),

/***/ "./src/environments/environment.ts":
/*!*****************************************!*\
  !*** ./src/environments/environment.ts ***!
  \*****************************************/
/*! exports provided: environment */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony export (binding) */ __webpack_require__.d(__webpack_exports__, "environment", function() { return environment; });
// This file can be replaced during build by using the `fileReplacements` array.
// `ng build --prod` replaces `environment.ts` with `environment.prod.ts`.
// The list of file replacements can be found in `angular.json`.
var environment = {
    production: false
};
/*
 * For easier debugging in development mode, you can import the following file
 * to ignore zone related error stack frames such as `zone.run`, `zoneDelegate.invokeTask`.
 *
 * This import should be commented out in production mode because it will have a negative impact
 * on performance if an error is thrown.
 */
// import 'zone.js/dist/zone-error';  // Included with Angular CLI.


/***/ }),

/***/ "./src/main.ts":
/*!*********************!*\
  !*** ./src/main.ts ***!
  \*********************/
/*! no exports provided */
/***/ (function(module, __webpack_exports__, __webpack_require__) {

"use strict";
__webpack_require__.r(__webpack_exports__);
/* harmony import */ var _angular_core__WEBPACK_IMPORTED_MODULE_0__ = __webpack_require__(/*! @angular/core */ "./node_modules/@angular/core/fesm5/core.js");
/* harmony import */ var _angular_platform_browser_dynamic__WEBPACK_IMPORTED_MODULE_1__ = __webpack_require__(/*! @angular/platform-browser-dynamic */ "./node_modules/@angular/platform-browser-dynamic/fesm5/platform-browser-dynamic.js");
/* harmony import */ var _app_app_module__WEBPACK_IMPORTED_MODULE_2__ = __webpack_require__(/*! ./app/app.module */ "./src/app/app.module.ts");
/* harmony import */ var _environments_environment__WEBPACK_IMPORTED_MODULE_3__ = __webpack_require__(/*! ./environments/environment */ "./src/environments/environment.ts");




if (_environments_environment__WEBPACK_IMPORTED_MODULE_3__["environment"].production) {
    Object(_angular_core__WEBPACK_IMPORTED_MODULE_0__["enableProdMode"])();
}
Object(_angular_platform_browser_dynamic__WEBPACK_IMPORTED_MODULE_1__["platformBrowserDynamic"])().bootstrapModule(_app_app_module__WEBPACK_IMPORTED_MODULE_2__["AppModule"])
    .catch(function (err) { return console.error(err); });


/***/ }),

/***/ 0:
/*!***************************!*\
  !*** multi ./src/main.ts ***!
  \***************************/
/*! no static exports found */
/***/ (function(module, exports, __webpack_require__) {

module.exports = __webpack_require__(/*! /Users/andrastoth/Developer/src/github.com/banzaicloud/uniform/src/main.ts */"./src/main.ts");


/***/ })

},[[0,"runtime","vendor"]]]);
//# sourceMappingURL=main.js.map