import { Routes } from '@angular/router';
import { LoginComponent } from './auth/login/login.component';
import { HomeComponent } from './home/home.component';
import { RepositoryDetailComponent } from './repositories/repository-detail/repository-detail.component';
import { RepositoryFormComponent } from './repositories/repository-form/repository-form.component';
import { RepositoryImportComponent } from './repositories/repository-import/repository-import.component';
import { authGuard } from './auth/auth.guard';

export const routes: Routes = [
  { path: '', redirectTo: '/home', pathMatch: 'full' },
  { path: 'login', component: LoginComponent },
  { path: 'home', component: HomeComponent, canActivate: [authGuard] },
  { path: 'repositories/new', component: RepositoryFormComponent, canActivate: [authGuard] },
  { path: 'repositories/import', component: RepositoryImportComponent, canActivate: [authGuard] },
  { path: 'repositories/:id', component: RepositoryDetailComponent, canActivate: [authGuard] },
  { path: '**', redirectTo: '/home' }
];
