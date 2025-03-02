import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { HttpClientModule } from '@angular/common/http';
import { RouterModule } from '@angular/router';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

import { AppComponent } from './app.component';
import { HomeComponent } from './home/home.component';
import { AuthModule } from './auth/auth.module';
import { ComponentsModule } from './components/components.module';
import { MaterialModule } from './material/material.module';
import { MarkdownModule } from './markdown/markdown.module';

@NgModule({
  declarations: [

  ],
  imports: [
    BrowserModule,
    HttpClientModule,
    RouterModule,
    FormsModule,
    ReactiveFormsModule,
    BrowserAnimationsModule,
    AuthModule,
    ComponentsModule,
    MaterialModule,
    MarkdownModule,
    AppComponent
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
