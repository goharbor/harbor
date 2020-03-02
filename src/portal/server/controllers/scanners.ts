import { Scanner } from '../../src/app/config/scanner/scanner';
import { Request, Response } from "express";


export function getScanner(req: Request, res: Response) {
  const scanners: Scanner[] =  [new Scanner(), new Scanner()];
  res.json(scanners);
}


