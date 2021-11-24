import { Component, OnInit } from '@angular/core';
import {Router} from '@angular/router';
import { BioairService} from '../../services/bioair.service';

@Component({
  selector: 'app-week-forecast',
  templateUrl: './week-forecast.component.html',
  styleUrls: ['./week-forecast.component.css']
})
export class WeekForecastComponent implements OnInit {

  progress_bar: boolean = false;
  allDays: any[] = [];
  constructor(
    private _bioAirService: BioairService,
    private router: Router
    ) { }



  ngOnInit(): void {
    this.progress_bar = true;
    this._bioAirService.getCarterasByUser().subscribe((res:any) =>{
      this.progress_bar = false;
      this.redondearValores(res)
      // this.todayInfo = res[0];
      // console.log(this.todayInfo)
      this.allDays = res;
      
      console.log(res);
    })
    
  }

  goToDetails(day:any, i:any){
    day['index'] = i+1
    this.router.navigate(['day-detail', day])
  }

  redondearValores(arr:any[]){
    var newArr = arr.map(x => {
      // console.log(x)
      x.ruido =  parseFloat(x.ruido.toFixed(2))
      x.uv = parseFloat(x.uv.toFixed(2))
      x.humedad = parseFloat(x.humedad.toFixed(2))
      x.presion = parseFloat(x.presion.toFixed(2))
      x.temperatura = parseFloat(x.temperatura.toFixed(2))
      x['descripcion'] = this.calcularTiempo(x.temperatura)
      return x;
    })

    console.log(newArr)
  }

  calcularTiempo(temp:any){
    if(temp <14){
      return 'Probablemente lluvioso'
    }else if(temp >= 14 && temp <17){
      return 'Probablemente nublado'
    }else if(temp >= 17 && temp <20){
      return 'Probablemente despejado'
    }else{
      return 'Probablemente soleado'
    }

  }

}
