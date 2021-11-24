import { Component, OnInit } from '@angular/core';
import { BioairService} from '../../services/bioair.service';

@Component({
  selector: 'app-inicio',
  templateUrl: './inicio.component.html',
  styleUrls: ['./inicio.component.css']
})
export class InicioComponent implements OnInit {

  TODAY_DATE: Date = new Date();
  MAX_DATE: Date = new Date();
  progress_bar: boolean = false;

  todayInfo: any;
  allDays: any[] = [];
  constructor(private _bioAirService: BioairService) { }

  ngOnInit(): void {
    this.progress_bar = true;
    this.MAX_DATE.setDate(this.TODAY_DATE.getUTCDate()+7)


    this._bioAirService.getCarterasByUser().subscribe((res:any) =>{
      this.progress_bar = false;
      console.log(this.todayInfo)
      this.redondearValores(res)
      this.todayInfo = res[0];
      localStorage.setItem('temp-hoy', this.todayInfo.temperatura.toFixed(2)+'°')
      this.allDays = res;
     
      
      // console.log(res);
    })
  }

 

  chooseDateWeather(e:any){
    console.log(e.value);
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
