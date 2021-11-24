import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { DayDetailComponent } from './components/day-detail/day-detail.component';
import { InicioComponent } from './components/inicio/inicio.component';
import { WeekForecastComponent } from './components/week-forecast/week-forecast.component';

const routes: Routes = [
    { path: '', redirectTo: 'inicio', pathMatch: 'full' },
    { path: 'inicio', component: InicioComponent},
    { path: 'week-forecast', component: WeekForecastComponent},
    { path: 'day-detail', component: DayDetailComponent},
];



@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
