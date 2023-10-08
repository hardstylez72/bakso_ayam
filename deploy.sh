#!/bin/bash

sudo docker build -f Dockerfile -t hardstylez72/bakso_ayam:v0.0.4 .
sudo docker push hardstylez72/bakso_ayam:v0.0.4
