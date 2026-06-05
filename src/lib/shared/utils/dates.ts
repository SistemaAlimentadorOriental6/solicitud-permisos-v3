import Holidays from 'date-holidays';

const MESES = [
  'enero', 'febrero', 'marzo', 'abril', 'mayo', 'junio',
  'julio', 'agosto', 'septiembre', 'octubre', 'noviembre', 'diciembre'
];

const DIAS = ['domingo', 'lunes', 'martes', 'miércoles', 'jueves', 'viernes', 'sábado'];

const hd = new Holidays('CO');

function getColombiaNow(): Date {
  const now = new Date();
  const str = now.toLocaleString('en-US', { timeZone: 'America/Bogota' });
  return new Date(str);
}

function isHoliday(date: Date): boolean {
  const holidays = hd.getHolidays(date.getFullYear());
  return holidays.some((h) => {
    const hDate = new Date(h.date);
    return hDate.getFullYear() === date.getFullYear() &&
           hDate.getMonth() === date.getMonth() &&
           hDate.getDate() === date.getDate();
  });
}

function getMondayOfWeek(date: Date): Date {
  const d = new Date(date);
  const dow = d.getDay();
  const normalized = dow === 0 ? 7 : dow;
  d.setDate(d.getDate() + (1 - normalized));
  d.setHours(0, 0, 0, 0);
  return d;
}

export function getWeekDates(forCalamidad: boolean = false): {
  dates: Array<{
    date: Date;
    dayName: string;
    dayNumber: string;
    monthName: string;
    year: string;
    shortDate: string;
    isHoliday: boolean;
    isToday?: boolean;
  }>;
  isExtemporaneous: boolean;
  weekLabel: string;
} {
  const now = getColombiaNow();
  const monday = getMondayOfWeek(now);
  const dow = now.getDay();
  const normalizedDow = dow === 0 ? 7 : dow;
  const hour = now.getHours();

  let weeksToAdd = 7;
  if (normalizedDow > 3 || (normalizedDow === 3 && hour >= 12)) {
    weeksToAdd = 14;
  }

  const startDate = new Date(monday);
  startDate.setDate(monday.getDate() + weeksToAdd);

  const isExtemporaneous = weeksToAdd === 14;

  const dates = [];
  let current = new Date(startDate);

  for (let i = 0; i < 7; i++) {
    dates.push({
      date: new Date(current),
      dayName: DIAS[current.getDay()],
      dayNumber: String(current.getDate()),
      monthName: MESES[current.getMonth()],
      year: String(current.getFullYear()),
      shortDate: `${String(current.getDate()).padStart(2, '0')}/${String(current.getMonth() + 1).padStart(2, '0')}`,
      isHoliday: isHoliday(current),
    });
    current.setDate(current.getDate() + 1);
  }

  const nextMonday = new Date(startDate);
  nextMonday.setDate(nextMonday.getDate() + 7);
  if (isHoliday(nextMonday)) {
    dates.push({
      date: new Date(nextMonday),
      dayName: DIAS[nextMonday.getDay()],
      dayNumber: String(nextMonday.getDate()),
      monthName: MESES[nextMonday.getMonth()],
      year: String(nextMonday.getFullYear()),
      shortDate: `${String(nextMonday.getDate()).padStart(2, '0')}/${String(nextMonday.getMonth() + 1).padStart(2, '0')}`,
      isHoliday: true,
    });
  }

  if (forCalamidad) {
    const today = new Date(now);
    today.setHours(0, 0, 0, 0);

    const alreadyInDates = dates.some(d => {
      const dDate = new Date(d.date);
      dDate.setHours(0, 0, 0, 0);
      return dDate.getTime() === today.getTime();
    });

    if (!alreadyInDates) {
      dates.unshift({
        date: today,
        dayName: DIAS[today.getDay()],
        dayNumber: String(today.getDate()),
        monthName: MESES[today.getMonth()],
        year: String(today.getFullYear()),
        shortDate: `${String(today.getDate()).padStart(2, '0')}/${String(today.getMonth() + 1).padStart(2, '0')}`,
        isHoliday: isHoliday(today),
        isToday: true,
      });
    }
  }

  const firstDate = dates[0];
  const lastDate = dates[dates.length - 1];
  const weekLabel = `${firstDate.dayNumber} ${firstDate.monthName.slice(0, 3)} - ${lastDate.dayNumber} ${lastDate.monthName.slice(0, 3)} ${lastDate.year}`;

  return { dates, isExtemporaneous, weekLabel };
}