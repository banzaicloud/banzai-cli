import { Component, ElementRef, ViewChild, OnInit } from '@angular/core';
import { MatSnackBar } from '@angular/material';
import { UniformService, CONTROL_TYPES, CONTROL_GROUP_TYPES, IFormFieldGroup, IFormField } from '@banzaicloud/uniform';
import { HttpClient, HttpHeaders, HttpErrorResponse } from '@angular/common/http';
import { map, catchError } from 'rxjs/operators';
import * as fp from 'lodash/fp';
import { Observable, throwError } from 'rxjs';

const reduceObject = fp.reduce.convert({ cap: false });
const mapObject = fp.map.convert({ cap: false });

// see https://codemirror.net/mode/index.html
import 'codemirror/mode/javascript/javascript';
import 'codemirror/mode/markdown/markdown';
import 'codemirror/mode/yaml/yaml';
// see https://codemirror.net/demo/placeholder.html
import 'codemirror/addon/display/placeholder';

@Component({
  selector: 'banzai-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss'],
})
export class AppComponent implements OnInit {
  rawGroups: any[];
  groups$: Observable<IFormFieldGroup[]>;
  values: { [key: string]: any };
  initialValue: any;
  error: string;

  @ViewChild('downloadLink')
  downloadLink: ElementRef;

  constructor(private http: HttpClient, private snackBar: MatSnackBar) {}

  ngOnInit() {
    this.groups$ = this.http
      .get<{ name: string; description?: string; link?: string; fields: any[] }[]>('/api/v1/form')
      .pipe(map((g) => UniformService.factory(g)))
      .pipe(
        catchError((err: HttpErrorResponse) => {
          this.error = `${err.message}\n\n${err.error || ''}`;
          return throwError(err);
        }),
      );
  }

  onSave(event) {
    const values = reduceObject((v, val, key) => fp.set(key, val, v), {})(event);

    this.http
      .post('/api/v1/form', values, {
        headers: new HttpHeaders({
          'Content-Type': 'application/json',
        }),
      })
      .pipe(
        catchError((err: HttpErrorResponse) => {
          this.error = `${err.message}\n\n${err.error || ''}`;
          return throwError(err);
        }),
      )
      .subscribe(() => {
        this.snackBar.open('Form values has been saved to config file', '⤬');
      });
  }
}
