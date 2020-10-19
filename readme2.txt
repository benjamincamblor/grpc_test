Ejecutar modulos en secuencia:

Primero ejecutar modulo de Camion hospedado en el host dist107.
Segundo ejecutar modulo de Finanzas hospedado en el host dist108.
Tercero ejecutar modulo de Logistica hospedado en el host dist106.
Cuarto ejecutar modulo de Clientes hospedado en el host dist105.

Para finalizar correctamente la ejecución del módulo de Camión y generar sus registros de despachos, presionar
'c' + Enter en su terminal correspondiente (dist107). Luego Control+C para detener al servidor completamente.

Para finalizar correctamente la ejecución del módulo de Finanzas y generar sus respectivos registros, presionar 
't'+Enter.

Cliente realiza comienza a realizar solicitudes de seguimiento automaticamente una vez transcurridos 10 segundos desde el inicio
de su ejecución. 

Para ejecutar un modulo, ingresar al host indicado y ejecutar el comando 'make run'. Cada módulo solicitara los parámetros
pertinentes según las especificaciones del enunciado. 
